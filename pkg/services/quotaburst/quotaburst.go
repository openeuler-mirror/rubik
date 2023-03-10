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
// Date: 2023-03-01
// Description: This file is used for quota burst

// Package  quotaburst is for Quota Burst
package quotaburst

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services/helper"
)

// Burst is used to control cpu burst
type Burst struct {
	helper.ServiceBase
	name string
}

type BurstFactory struct {
	ObjName string
}

func (i BurstFactory) Name() string {
	return "BurstFactory"
}

func (i BurstFactory) NewObj() (interface{}, error) {
	return &Burst{name: i.ObjName}, nil
}

// ID returns the module name
func (b *Burst) ID() string {
	return b.name
}

// AddFunc implement add function when pod is added in k8s
func (conf *Burst) AddFunc(podInfo *typedef.PodInfo) error {
	return setPodQuotaBurst(podInfo)
}

// UpdateFunc implement update function when pod info is changed
func (conf *Burst) UpdateFunc(oldPod, newPod *typedef.PodInfo) error {
	if oldPod.Annotations[constant.QuotaBurstAnnotationKey] == newPod.Annotations[constant.QuotaBurstAnnotationKey] {
		return nil
	}
	return setPodQuotaBurst(newPod)
}

// DeleteFunc implement delete function when pod is deleted by k8s
func (conf *Burst) DeleteFunc(podInfo *typedef.PodInfo) error {
	return nil
}

// PreStart is the pre-start action
func (conf *Burst) PreStart(viewer api.Viewer) error {
	pods := viewer.ListPodsWithOptions()
	for _, pod := range pods {
		if err := setPodQuotaBurst(pod); err != nil {
			log.Errorf("error prestart cont %v: %v", pod.Name, err)
		}
	}
	return nil
}

func setPodQuotaBurst(podInfo *typedef.PodInfo) error {
	if podInfo.Annotations[constant.QuotaBurstAnnotationKey] == "" {
		return nil
	}
	burst, err := parseQuotaBurst(podInfo)
	if err != nil {
		return err
	}
	var podBurst int64 = 0
	const subsys = "cpu"
	// 1. Try to write container burst value firstly
	for _, c := range podInfo.IDContainersMap {
		cgpath := cgroup.AbsoluteCgroupPath(subsys, c.CgroupPath, "")
		if err := setQuotaBurst(burst, cgpath); err != nil {
			log.Errorf("set container quota burst failed: %v", err)
			continue
		}
		/*
			Only when the burst value of the container is successfully set,
			the burst value of the pod will be accumulated.
			Ensure that Pod data must be written successfully
		*/
		podBurst += burst
	}
	// 2. Try to write pod burst value
	podPath := cgroup.AbsoluteCgroupPath(subsys, podInfo.CgroupPath, "")
	if err := setQuotaBurst(podBurst, podPath); err != nil {
		log.Errorf("set pod quota burst failed: %v", err)
	}
	return nil
}

func setQuotaBurst(burst int64, cgpath string) error {
	const burstFileName = "cpu.cfs_burst_us"
	fpath := filepath.Join(cgpath, burstFileName)
	// check whether cgroup support cpu burst
	if _, err := os.Stat(fpath); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("quota-burst path=%v missing", fpath)
	}
	if err := matchQuota(burst, cgpath); err != nil {
		return err
	}
	// try to write cfs_burst_us
	if err := ioutil.WriteFile(fpath, []byte(util.FormatInt64(burst)), constant.DefaultFileMode); err != nil {
		return fmt.Errorf("quota-burst path=%v setting failed: %v", fpath, err)
	}
	log.Infof("quota-burst path=%v setting success", fpath)
	return nil
}

func matchQuota(burst int64, cgpath string) error {
	const (
		cpuPeriodFileName = "cpu.cfs_period_us"
		cpuQuotaFileName  = "cpu.cfs_quota_us"
	)
	quotaStr, err := util.ReadSmallFile(filepath.Join(cgpath, cpuQuotaFileName))
	if err != nil {
		return fmt.Errorf("fail to read cfs.cpu_quota_us: %v", err)
	}
	quota, err := util.ParseInt64(strings.TrimSpace(string(quotaStr)))
	if err != nil {
		return fmt.Errorf("fail to parse quota as int64: %v", err)
	}

	periodStr, err := util.ReadSmallFile(filepath.Join(cgpath, cpuPeriodFileName))
	if err != nil {
		return fmt.Errorf("fail to read cfs.cpu_period_us: %v", err)
	}
	period, err := util.ParseInt64(strings.TrimSpace(string(periodStr)))
	if err != nil {
		return fmt.Errorf("fail to parse period as int64: %v", err)
	}

	/*
		The current pod has been allowed to use all cores, usually there are two situations:
		1.the pod quota is -1 (in this case, there must be a container with a quota of -1)
		2.the pod quota exceeds the maximum value (the cumulative quota value of all containers
			exceeds the maximum value)
	*/
	maxQuota := period * int64(runtime.NumCPU())
	if quota >= maxQuota {
		return fmt.Errorf("burst fail when quota exceed the maxQuota")
	}
	/*
		All containers under the pod have set cpulimit, and the cumulative value is less than the maximum core.
		At this time, the quota of the pod should be the accumulated value of the quota of all pods.
		If the burst value of the container is set successfully, then the burst value of the Pod
		must be set successfully
	*/
	if quota < burst {
		return fmt.Errorf("burst should be less than or equal to quota")
	}
	return nil
}

// parseQuotaBurst checks CPU quota burst annotation value.
func parseQuotaBurst(pod *typedef.PodInfo) (int64, error) {
	const invalidVal int64 = -1
	val, err := util.ParseInt64(pod.Annotations[constant.QuotaBurstAnnotationKey])
	if err != nil {
		return invalidVal, err
	}

	if val < 0 {
		return invalidVal, fmt.Errorf("quota burst value should be positive")
	}
	return val, nil
}
