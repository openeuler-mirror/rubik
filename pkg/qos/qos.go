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
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	corev1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
)

var (
	// SupportCgroupTypes are supported cgroup types for qos setting
	SupportCgroupTypes = []string{"cpu", "memory"}
)

// PodInfo is struct of each pod's info
type PodInfo struct {
	api.PodQoS
	PodID      string
	CgroupRoot string
	FullPath   map[string]string
	Ctx        context.Context
}

// NewPodInfo is constructor of struct PodInfo
func NewPodInfo(ctx context.Context, podID string, cgmnt string, req api.PodQoS) (*PodInfo, error) {
	if len(podID) > constant.MaxPodIDLen {
		return nil, errors.Errorf("Pod id too long")
	}
	pod := PodInfo{
		PodQoS: api.PodQoS{
			CgroupPath: req.CgroupPath,
			QosLevel:   req.QosLevel,
		},
		PodID:      podID,
		CgroupRoot: cgmnt,
		Ctx:        ctx,
	}
	if err := checkQosLevel(pod.QosLevel); err != nil {
		return nil, err
	}
	if err := pod.initCgroupPath(); err != nil {
		return nil, err
	}

	return &pod, nil
}

// BuildOfflinePodInfo build offline pod information
func BuildOfflinePodInfo(pod *corev1.Pod) (*PodInfo, error) {
	podQos := api.PodQoS{
		CgroupPath: util.GetPodCgroupPath(pod),
		QosLevel:   -1,
	}
	podInfo, err := NewPodInfo(context.Background(), string(pod.UID), config.CgroupRoot, podQos)
	if err != nil {
		return nil, err
	}

	return podInfo, nil
}

func getQosLevel(root, file string) (int, error) {
	var (
		qosLevel int
		rootQos  []byte
		err      error
	)

	rootQos, err = util.ReadSmallFile(filepath.Join(root, file)) // nolint
	if err != nil {
		return constant.ErrCodeFailed, errors.Errorf("Getting root qos level failed: %v", err)
	}
	// walk through all sub paths
	if err = filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if f != nil && f.IsDir() {
			cgFilePath, err := securejoin.SecureJoin(path, file)
			if err != nil {
				return errors.Errorf("Join path failed: %v", err)
			}
			data, err := util.ReadSmallFile(filepath.Clean(cgFilePath))
			if err != nil {
				return errors.Errorf("Getting qos level failed: %v", err)
			}
			if strings.Compare(string(data), string(rootQos)) != 0 {
				return errors.Errorf("Qos differs")
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

func checkQosLevel(qosLevel int) error {
	if qosLevel >= constant.MinLevel.Int() && qosLevel <= constant.MaxLevel.Int() {
		return nil
	}

	return errors.Errorf("invalid qos level number %d, should be 0 or -1", qosLevel)
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

// SetQos is used for setting pod's qos level following it's cgroup path
func (pod *PodInfo) SetQos() error {
	ctx := pod.Ctx
	log.WithCtx(ctx).Logf("Setting level=%d for pod %s", pod.QosLevel, pod.PodID)
	if pod.FullPath == nil {
		return errors.Errorf("Empty cgroup path of pod %s", pod.PodID)
	}

	for kind, cgPath := range pod.FullPath {
		switch kind {
		case "cpu":
			if err := setQosLevel(cgPath, constant.CPUCgroupFileName, pod.QosLevel); err != nil {
				return err
			}
		case "memory":
			if err := setQosLevel(cgPath, constant.MemoryCgroupFileName, pod.QosLevel); err != nil {
				return err
			}
		}
	}

	log.WithCtx(ctx).Logf("Setting level=%d for pod %s OK", pod.QosLevel, pod.PodID)
	return nil
}

// ValidateQos is used for checking pod's qos level if equal to the value it should be set up to
func (pod *PodInfo) ValidateQos() error {
	var (
		cpuInfo, memInfo int
		err              error
	)
	ctx := pod.Ctx

	log.WithCtx(ctx).Logf("Checking level=%d for pod %s", pod.QosLevel, pod.PodID)

	for kind, cgPath := range pod.FullPath {
		switch kind {
		case "cpu":
			if cpuInfo, err = getQosLevel(cgPath, constant.CPUCgroupFileName); err != nil {
				return errors.Errorf("read %s for pod %q failed: %v", constant.CPUCgroupFileName, pod.PodID, err)
			}
		case "memory":
			if memInfo, err = getQosLevel(cgPath, constant.MemoryCgroupFileName); err != nil {
				return errors.Errorf("read %s for pod %q failed: %v", constant.MemoryCgroupFileName, pod.PodID, err)
			}
		}
	}

	if (cpuInfo != pod.QosLevel) || (memInfo != pod.QosLevel) {
		return errors.Errorf("checking level=%d for pod %s failed", pod.QosLevel, pod.PodID)
	}

	log.WithCtx(ctx).Logf("Checking level=%d for pod %s OK", pod.QosLevel, pod.PodID)

	return nil
}

// initCgroupPath return pod's cgroup full path
func (pod *PodInfo) initCgroupPath() error {
	if pod.CgroupRoot == "" {
		pod.CgroupRoot = constant.DefaultCgroupRoot
	}
	cgroupMap := make(map[string]string, len(SupportCgroupTypes))
	for _, kind := range SupportCgroupTypes {
		if err := checkCgroupPath(pod.CgroupPath); err != nil {
			return err
		}
		fullPath := filepath.Join(pod.CgroupRoot, kind, pod.CgroupPath)
		if len(fullPath) > constant.MaxCgroupPathLen {
			return errors.Errorf("length of cgroup path exceeds max limit %d", constant.MaxCgroupPathLen)
		}
		cgroupMap[kind] = fullPath
	}

	pod.FullPath = cgroupMap

	return nil
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
