// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2023-01-05
// Description: This file defines ContainerInfo

// Package typedef defines core struct and methods for rubik
package typedef

import (
	"fmt"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// ContainerInfo contains the interested information of container
type ContainerInfo struct {
	cgroup.Hierarchy
	Name             string      `json:"name"`
	ID               string      `json:"id"`
	RequestResources ResourceMap `json:"requests,omitempty"`
	LimitResources   ResourceMap `json:"limits,omitempty"`
	PodSandboxId     string      `json:"podisandid,omitempty"` // id of the sandbox which can uniquely determine a pod
}

// DeepCopy returns deepcopy object.
func (cont *ContainerInfo) DeepCopy() *ContainerInfo {
	copyObject := *cont
	copyObject.LimitResources = cont.LimitResources.DeepCopy()
	copyObject.RequestResources = cont.RequestResources.DeepCopy()
	return &copyObject
}

type ContainerConfig struct {
	rawCont       *RawContainer
	nriCont       *NRIRawContainer
	request       ResourceMap
	limit         ResourceMap
	podCgroupPath string
}

type ConfigOpt func(b *ContainerConfig)

func WithRawContainer(cont *RawContainer) ConfigOpt {
	return func(conf *ContainerConfig) {
		conf.rawCont = cont
	}
}

func WithNRIContainer(cont *NRIRawContainer) ConfigOpt {
	return func(conf *ContainerConfig) {
		conf.nriCont = cont
	}
}

func WithRequest(req ResourceMap) ConfigOpt {
	return func(conf *ContainerConfig) {
		conf.request = req
	}
}

func WithLimit(limit ResourceMap) ConfigOpt {
	return func(conf *ContainerConfig) {
		conf.limit = limit
	}
}

func WithPodCgroup(path string) ConfigOpt {
	return func(conf *ContainerConfig) {
		conf.podCgroupPath = path
	}
}

func NewContainerInfo(opts ...ConfigOpt) *ContainerInfo {
	var (
		conf = &ContainerConfig{}
		ci   = &ContainerInfo{}
	)
	for _, opt := range opts {
		opt(conf)
	}

	if err := fromRawContainer(ci, conf.rawCont); err != nil {
		fmt.Printf("failed to parse raw container: %v", err)
	}
	fromNRIContainer(ci, conf.nriCont)
	fromPodCgroupPath(ci, conf.podCgroupPath)

	if conf.request != nil {
		ci.RequestResources = conf.request
	}

	if conf.limit != nil {
		ci.LimitResources = conf.limit
	}

	return ci
}

func fromRawContainer(ci *ContainerInfo, rawCont *RawContainer) error {
	if rawCont == nil {
		return nil
	}
	requests, limits := rawCont.GetResourceMaps()
	id, err := rawCont.GetRealContainerID()
	if err != nil {
		return fmt.Errorf("failed to parse container ID: %v", err)
	}
	if id == "" {
		return fmt.Errorf("empty container id")
	}

	ci.Name = rawCont.status.Name
	ci.ID = id
	ci.RequestResources = requests
	ci.LimitResources = limits
	return nil
}

// convert NRIRawContainer structure to ContainerInfo structure
func fromNRIContainer(ci *ContainerInfo, nriCont *NRIRawContainer) {
	if nriCont == nil {
		return
	}
	ci.ID = nriCont.Id
	ci.Hierarchy = cgroup.Hierarchy{
		Path: cgroup.GetNRIContainerCgroupPath(nriCont.Linux.GetCgroupsPath()),
	}
	ci.Name = nriCont.Name
	ci.PodSandboxId = nriCont.PodSandboxId
}

func fromPodCgroupPath(ci *ContainerInfo, podCgroupPath string) {
	if podCgroupPath == "" {
		return
	}
	// TODO : don't need to judge cgroup driver
	var prefix = containerEngineScopes[currentContainerEngines]
	if cgroup.Type() == constant.CgroupDriverCgroupfs && currentContainerEngines != CRIO {
		prefix = ""
	}
	ci.Hierarchy = cgroup.Hierarchy{Path: cgroup.ConcatContainerCgroup(podCgroupPath, prefix, ci.ID)}
}
