// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: HuangYuqing
// Create: 2022-10-26
// Description: iocost setting for pods.

// Package iocost is for iocost.
package iocost

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/pkg/errors"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
	"isula.org/rubik/pkg/util"
)

const (
	iocostModelFile  = "blkio.cost.model"
	iocostWeightFile = "blkio.cost.weight"
	iocostQosFile    = "blkio.cost.qos"
	wbBlkioinoFile   = "memory.wb_blkio_ino"
	blkSubName       = "blkio"
	memSubName       = "memory"
	offlineWeight    = 10
	onlineWeight     = 1000
	paramMaxLen      = 512
	devNoMax         = 256
	scale            = 10
	sysDevBlock      = "/sys/dev/block"
)

var (
	hwSupport    = false
	iocostEnable = false
)

// HwSupport tell if the os support iocost.
func HwSupport() bool {
	return hwSupport
}

func init() {
	qosFile := filepath.Join(constant.DefaultCgroupRoot, blkSubName, iocostQosFile)
	if util.PathExist(qosFile) {
		hwSupport = true
	}
}

// SetIOcostEnable set iocost disable or enable
func SetIOcostEnable(status bool) {
	iocostEnable = status
}

// ConfigIOcost for config iocost in cgroup v1.
func ConfigIOcost(iocostConfigArray []config.IOcostConfig) error {
	if !iocostEnable {
		return errors.Errorf("iocost feature is disable")
	}

	if err := clearIOcost(); err != nil {
		log.Infof(err.Error())
	}

	for _, iocostConfig := range iocostConfigArray {
		if !iocostConfig.Enable {
			// notice: dev's iocost is disable by clearIOcost
			continue
		}

		devno, err := getBlkDeviceNo(iocostConfig.Dev)
		if err != nil {
			log.Errorf(err.Error())
			continue
		}

		if iocostConfig.Model == "linear" {
			if err := configLinearModel(iocostConfig.Param, devno); err != nil {
				log.Errorf(err.Error())
				continue
			}
		} else {
			log.Errorf("curent rubik not support non-linear model")
			continue
		}

		if err := configQos(true, devno); err != nil {
			log.Errorf(err.Error())
			continue
		}
	}
	return nil
}

// SetPodWeight set pod weight
func SetPodWeight(pod *typedef.PodInfo) error {
	if !iocostEnable {
		return errors.Errorf("iocost feature is disable")
	}
	weightFile := filepath.Join(pod.CgroupRoot,
		blkSubName, pod.CgroupPath, iocostWeightFile)
	if err := configWeight(pod.Offline, weightFile); err != nil {
		return err
	}
	if err := bindMemcgBlkio(pod.Containers); err != nil {
		return err
	}
	return nil
}

// ShutDown for clear iocost if feature is enable.
func ShutDown() error {
	if !iocostEnable {
		return errors.Errorf("iocost feature is disable")
	}
	if err := clearIOcost(); err != nil {
		return err
	}
	return nil
}

func getBlkDeviceNo(devName string) (string, error) {
	devPath := filepath.Join("/dev", devName)
	fi, err := os.Stat(devPath)
	if err != nil {
		return "", errors.Errorf("stat %s failed with error: %v", devName, err)
	}

	if fi.Mode()&os.ModeDevice == 0 {
		return "", errors.Errorf("%s is not a device", devName)
	}

	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", errors.Errorf("failed to get Sys(), %v has type %v", devName, st)
	}

	devno := st.Rdev
	major, minor := devno/devNoMax, devno%devNoMax
	return fmt.Sprintf("%v:%v", major, minor), nil
}

func configWeight(offline bool, file string) error {
	var weight uint64 = offlineWeight
	if !offline {
		weight = onlineWeight
	}
	return writeIOcost(file, strconv.FormatUint(weight, scale))
}

func configQos(enable bool, devno string) error {
	t := 0
	if enable {
		t = 1
	}
	qosStr := fmt.Sprintf("%v enable=%v ctrl=user min=100.00 max=100.00", devno, t)
	filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostQosFile)
	return writeIOcost(filePath, qosStr)
}

func configLinearModel(linearModelParam config.Param, devno string) error {
	if linearModelParam.Rbps <= 0 || linearModelParam.Rseqiops <= 0 || linearModelParam.Rrandiops <= 0 ||
		linearModelParam.Wbps <= 0 || linearModelParam.Wseqiops <= 0 || linearModelParam.Wrandiops <= 0 {
		return errors.Errorf("invalid iocost.params, the value must not 0")
	}
	paramStr := fmt.Sprintf("%v rbps=%v rseqiops=%v rrandiops=%v wbps=%v wseqiops=%v wrandiops=%v",
		devno,
		linearModelParam.Rbps, linearModelParam.Rseqiops, linearModelParam.Rrandiops,
		linearModelParam.Wbps, linearModelParam.Wseqiops, linearModelParam.Wrandiops)
	filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostModelFile)
	return writeIOcost(filePath, paramStr)
}

func bindMemcgBlkio(containers map[string]*typedef.ContainerInfo) error {
	for _, container := range containers {
		memPath := container.CgroupPath(memSubName)
		blkPath := container.CgroupPath(blkSubName)
		ino, err := getDirInode(blkPath)
		if err != nil {
			log.Errorf("get director:%v, inode err:%v", blkPath, err.Error())
			continue
		}
		wbBlkFile := filepath.Join(memPath, wbBlkioinoFile)
		if err := writeIOcost(wbBlkFile, strconv.FormatUint(ino, scale)); err != nil {
			log.Errorf("write file %v err:%v", wbBlkFile, err.Error())
			continue
		}
	}
	return nil
}

func getDirInode(file string) (uint64, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, errors.Errorf("failed to get Sys(), %v has type %v", file, st)
	}
	return st.Ino, nil
}

func clearIOcost() error {
	qosFilePath := filepath.Join(config.CgroupRoot, blkSubName, iocostQosFile)
	qosParamByte, err := ioutil.ReadFile(qosFilePath)
	if err != nil {
		return errors.Errorf("read file:%v failed, err:%v", qosFilePath, err.Error())
	}

	if len(qosParamByte) == 0 {
		return errors.Errorf("read file:%v is empty", qosFilePath)
	}

	qosParams := strings.Split(string(qosParamByte), "\n")
	for _, param := range qosParams {
		paramList := strings.FieldsFunc(param, unicode.IsSpace)
		if len(paramList) != 0 {
			if err := configQos(false, paramList[0]); err != nil {
				return errors.Errorf("write file:%v failed, err:%v", qosFilePath, err.Error())
			}
		}
	}
	return nil
}

func writeIOcost(file, param string) error {
	if len(param) > paramMaxLen {
		return errors.Errorf("param size exceeds %v", paramMaxLen)
	}
	if !util.PathExist(file) {
		return errors.Errorf("path %v not exist, maybe iocost is unsupport", file)
	}
	err := ioutil.WriteFile(file, []byte(param), constant.DefaultFileMode)
	return err
}
