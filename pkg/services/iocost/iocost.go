// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2023-03-11
// Description: This file is used to implement iocost

// Package iocost
package iocost

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services/helper"
)

const (
	blkcgRootDir  = "blkio"
	memcgRootDir  = "memory"
	offlineWeight = 10
	onlineWeight  = 1000
	scale         = 10
)

// LinearParam for linear model
type LinearParam struct {
	Rbps      int64 `json:"rbps,omitempty"`
	Rseqiops  int64 `json:"rseqiops,omitempty"`
	Rrandiops int64 `json:"rrandiops,omitempty"`
	Wbps      int64 `json:"wbps,omitempty"`
	Wseqiops  int64 `json:"wseqiops,omitempty"`
	Wrandiops int64 `json:"wrandiops,omitempty"`
}

// IOCostConfig define iocost for node
type IOCostConfig struct {
	Dev   string      `json:"dev,omitempty"`
	Model string      `json:"model,omitempty"`
	Param LinearParam `json:"param,omitempty"`
}

// NodeConfig define the config of node, include iocost
type NodeConfig struct {
	NodeName     string         `json:"nodeName,omitempty"`
	IOCostConfig []IOCostConfig `json:"config,omitempty"`
}

// IOCost for iocost class
type IOCost struct {
	helper.ServiceBase
}

var (
	nodeName string
)

// IOCostFactory is the factory of IOCost.
type IOCostFactory struct {
	ObjName string
}

// Name to get the IOCost factory name.
func (i IOCostFactory) Name() string {
	return "IOCostFactory"
}

// NewObj to create object of IOCost.
func (i IOCostFactory) NewObj() (interface{}, error) {
	if ioCostSupport() {
		nodeName = os.Getenv(constant.NodeNameEnvKey)
		return &IOCost{ServiceBase: helper.ServiceBase{Name: i.ObjName}}, nil
	}
	return nil, fmt.Errorf("this machine not support iocost")
}

// ioCostSupport tell if the os support iocost.
func ioCostSupport() bool {
	cmdLine, err := util.ReadSmallFile("/proc/cmdline")
	if err != nil {
		log.Warnf("get /pro/cmdline error:%v", err)
		return false
	}

	if !strings.Contains(string(cmdLine), "cgroup1_writeback") {
		log.Warnf("current machine does not support writeback, please add 'cgroup1_writeback' to cmdline")
		return false
	}

	qosFile := cgroup.AbsoluteCgroupPath(blkcgRootDir, iocostQosFile)
	modelFile := cgroup.AbsoluteCgroupPath(blkcgRootDir, iocostModelFile)
	return util.PathExist(qosFile) && util.PathExist(modelFile)
}

// SetConfig to config nodeConfig configure
func (io *IOCost) SetConfig(f helper.ConfigHandler) error {
	if f == nil {
		return fmt.Errorf("no config handler function callback")
	}

	var nodeConfigs []NodeConfig
	if err := f(io.Name, &nodeConfigs); err != nil {
		return err
	}

	var nodeConfig *NodeConfig
	for _, config := range nodeConfigs {
		if config.NodeName == nodeName {
			nodeConfig = &config
			break
		}
		if config.NodeName == "global" {
			nodeConfig = &config
		}
	}
	return io.loadConfig(nodeConfig)
}

func (io *IOCost) loadConfig(nodeConfig *NodeConfig) error {
	// ensure that previous configuration is cleared.
	if err := io.clearIOCost(); err != nil {
		log.Errorf("failed to clear iocost:%v", err)
		return err
	}

	// no config, return
	if nodeConfig == nil {
		log.Warnf("no matching node exist:%v", nodeName)
		return nil
	}

	if err := io.configIOCost(nodeConfig.IOCostConfig); err != nil {
		if err2 := io.clearIOCost(); err2 != nil {
			log.Errorf("clear iocost failed:%v", err2)
		}
		return err
	}
	return nil
}

// PreStart is the pre-start action
func (io *IOCost) PreStart(viewer api.Viewer) error {
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	return io.dealExistedPods(viewer)
}

// Terminate is the terminating action
func (b *IOCost) Terminate(viewer api.Viewer) error {
	if err := b.clearIOCost(); err != nil {
		return err
	}
	return nil
}

func (b *IOCost) dealExistedPods(viewer api.Viewer) error {
	pods := viewer.ListPodsWithOptions()
	for _, pod := range pods {
		if err := b.configPodIOCostWeight(pod); err != nil {
			log.Errorf("config pod iocost failed, err:%v", err)
		}
	}
	return nil
}

// AddPod to deal the event of adding a pod.
func (b *IOCost) AddPod(podInfo *typedef.PodInfo) error {
	return b.configPodIOCostWeight(podInfo)
}

// UpdatePod to deal the pod update event.
func (b *IOCost) UpdatePod(old, new *typedef.PodInfo) error {
	return b.configPodIOCostWeight(new)
}

// DeletePod to deal the pod deletion event.
func (b *IOCost) DeletePod(podInfo *typedef.PodInfo) error {
	return nil
}

func (b *IOCost) configIOCost(configs []IOCostConfig) error {
	for _, config := range configs {
		devno, err := getBlkDeviceNo(config.Dev)
		if err != nil {
			return err
		}
		if config.Model == "linear" {
			if err := ConfigIOCostModel(devno, config.Param); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("non-linear models are not supported")
		}

		if err := ConfigIOCostQoS(devno, true); err != nil {
			return err
		}
	}
	return nil
}

// clearIOCost used to disable all iocost
func (b *IOCost) clearIOCost() error {
	qosbytes, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostQosFile)
	if err != nil {
		return err
	}

	if len(qosbytes) == 0 {
		return nil
	}

	qosParams := strings.Split(string(qosbytes), "\n")
	for _, qosParam := range qosParams {
		words := strings.FieldsFunc(qosParam, unicode.IsSpace)
		if len(words) != 0 {
			if err := ConfigIOCostQoS(words[0], false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *IOCost) configPodIOCostWeight(podInfo *typedef.PodInfo) error {
	var weight uint64 = offlineWeight
	if podInfo.Online() {
		weight = onlineWeight
	}
	return ConfigPodIOCostWeight(podInfo.Path, weight)
}
