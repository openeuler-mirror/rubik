// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2021-04-17
// Description: QoS setting for pods

package qos

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/pkg/errors"

	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
	"isula.org/rubik/pkg/util"
)

// SupportCgroupTypes are supported cgroup types for qos setting
var SupportCgroupTypes = []string{"cpu", "memory"}

// SetQosLevel set pod qos_level
func SetQosLevel(pod *typedef.PodInfo) error {
	if err := setQos(pod); err != nil {
		return errors.Errorf("set qos for pod %s(%s) error: %v", pod.Name, pod.UID, err)
	}
	if err := validateQos(pod); err != nil {
		return errors.Errorf("validate qos for pod %s(%s) error: %v", pod.Name, pod.UID, err)
	}

	log.Logf("Set pod %s(UID=%s, offline=%v) qos level OK", pod.Name, pod.UID, pod.Offline)
	return nil
}

func UpdateQosLevel(pod *typedef.PodInfo) error {
	if err := validateQos(pod); err != nil {
		log.Logf("Checking pod %s(%s) value failed: %v, reset it", err, pod.Name, pod.UID)
		if err := setQos(pod); err != nil {
			return errors.Errorf("set qos for pod %s(%s) error: %v", pod.Name, pod.UID, err)
		}
	}

	return nil
}

// setQos is used for setting pod's qos level following it's cgroup path
func setQos(pod *typedef.PodInfo) error {
	if len(pod.UID) > constant.MaxPodIDLen {
		return errors.Errorf("Pod id too long")
	}

	// default qos_level is online, no need to set online pod qos_level
	if !pod.Offline {
		log.Logf("Set level=%v for pod %s(%s)", constant.MaxLevel, pod.Name, pod.UID)
		return nil
	}
	log.Logf("Set level=%v for pod %s(%s)", constant.MinLevel, pod.Name, pod.UID)

	cgroupMap, err := initCgroupPath(pod.CgroupRoot, pod.CgroupPath)
	if err != nil {
		return err
	}

	for kind, cgPath := range cgroupMap {
		switch kind {
		case "cpu":
			if err := setQosLevel(cgPath, constant.CPUCgroupFileName, int(constant.MinLevel)); err != nil {
				return err
			}
		case "memory":
			if err := setQosLevel(cgPath, constant.MemoryCgroupFileName, int(constant.MinLevel)); err != nil {
				return err
			}
		}
	}

	return nil
}

func setQosLevel(root, file string, target int) error {
	if !util.IsDirectory(root) {
		return errors.Errorf("Invalid cgroup path %q", root)
	}
	if old, err := getQosLevel(root, file); err == nil && target > old {
		return errors.Errorf("Not support change qos level from low to high")
	}
	// walk through all sub paths
	if err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if f != nil && f.IsDir() {
			cgFilePath, err := securejoin.SecureJoin(path, file)
			if err != nil {
				return errors.Errorf("Join path failed for %s and %s: %v", path, file, err)
			}
			if err = ioutil.WriteFile(cgFilePath, []byte(strconv.Itoa(target)),
				constant.DefaultFileMode); err != nil {
				return errors.Errorf("Setting qos level failed for %s=%d: %v", cgFilePath, target, err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// validateQos is used for checking pod's qos level if equal to the value it should be set up to
func validateQos(pod *typedef.PodInfo) error {
	var (
		cpuInfo, memInfo int
		err              error
		qosLevel         int
	)

	if !pod.Offline {
		qosLevel = int(constant.MaxLevel)
	} else {
		qosLevel = int(constant.MinLevel)
	}

	cgroupMap, err := initCgroupPath(pod.CgroupRoot, pod.CgroupPath)
	if err != nil {
		return err
	}
	for kind, cgPath := range cgroupMap {
		switch kind {
		case "cpu":
			if cpuInfo, err = getQosLevel(cgPath, constant.CPUCgroupFileName); err != nil {
				return errors.Errorf("read %s failed: %v", constant.CPUCgroupFileName, err)
			}
		case "memory":
			if memInfo, err = getQosLevel(cgPath, constant.MemoryCgroupFileName); err != nil {
				return errors.Errorf("read %s failed: %v", constant.MemoryCgroupFileName, err)
			}
		}
	}

	if (cpuInfo != qosLevel) || (memInfo != qosLevel) {
		return errors.Errorf("check level failed")
	}

	return nil
}

func getQosLevel(root, file string) (int, error) {
	var (
		qosLevel int
		rootQos  []byte
		err      error
	)

	rootQos, err = util.ReadSmallFile(filepath.Join(root, file)) // nolint
	if err != nil {
		return constant.ErrCodeFailed, errors.Errorf("get root qos level failed: %v", err)
	}
	// walk through all sub paths
	if err = filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if f != nil && f.IsDir() {
			cgFilePath, err := securejoin.SecureJoin(path, file)
			if err != nil {
				return errors.Errorf("join path failed: %v", err)
			}
			data, err := util.ReadSmallFile(filepath.Clean(cgFilePath))
			if err != nil {
				return errors.Errorf("get qos level failed: %v", err)
			}
			if strings.Compare(string(data), string(rootQos)) != 0 {
				return errors.Errorf("qos differs")
			}
		}
		return nil
	}); err != nil {
		return constant.ErrCodeFailed, err
	}
	qosLevel, err = strconv.Atoi(strings.TrimSpace(string(rootQos)))
	if err != nil {
		return constant.ErrCodeFailed, err
	}

	return qosLevel, nil
}

// initCgroupPath return pod's cgroup full path
func initCgroupPath(cgroupRoot, cgroupPath string) (map[string]string, error) {
	if cgroupRoot == "" {
		cgroupRoot = constant.DefaultCgroupRoot
	}
	cgroupMap := make(map[string]string, len(SupportCgroupTypes))
	for _, kind := range SupportCgroupTypes {
		if err := checkCgroupPath(cgroupPath); err != nil {
			return nil, err
		}
		fullPath := filepath.Join(cgroupRoot, kind, cgroupPath)
		if len(fullPath) > constant.MaxCgroupPathLen {
			return nil, errors.Errorf("length of cgroup path exceeds max limit %d", constant.MaxCgroupPathLen)
		}
		cgroupMap[kind] = fullPath
	}

	return cgroupMap, nil
}

func checkCgroupPath(path string) error {
	pathPrefix, blacklist := "kubepods", []string{"kubepods", "kubepods/besteffort", "kubepods/burstable"}
	cPath := filepath.Clean(path)

	if !strings.HasPrefix(cPath, pathPrefix) {
		return errors.Errorf("invalid cgroup path %v, should start with %v", path, pathPrefix)
	}

	for _, invalidPath := range blacklist {
		if cPath == invalidPath {
			return errors.Errorf("invalid cgroup path %v, without podID", path)
		}
	}

	return nil
}
