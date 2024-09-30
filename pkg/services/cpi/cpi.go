// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Kelu Ye
// Date: 2024-09-19
// Description: This file provides the method for detecting and handling the interference of offline tasks on online tasks in the CPU Interference Detection Service.

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services/helper"
)

// CpiFactory is the CPI factory class
type CpiFactory struct {
	ObjName string
}

// Name returns the factory class name
func (i CpiFactory) Name() string {
	return "CpiFactory"
}

// NewObj returns a CPI object
func (i CpiFactory) NewObj() (interface{}, error) {
	return newCpiService(i.ObjName), nil
}

// CpiService manages CPU quota data for all online and offline pods on the current node.
type CpiService struct {
	helper.ServiceBase
	onlineTasks  map[string]*podStatus
	offlineTasks map[string]*podStatus
	Viewer       api.Viewer
	onlineMutex  sync.RWMutex
	offlineMutex sync.RWMutex
}

func (service *CpiService) addPod(pod *typedef.PodInfo) {
	switch pod.Annotations[constant.CpiAnnotationKey] {
	case "online":
		service.onlineMutex.Lock()
		service.onlineTasks[pod.UID], _ = newPodStatus(pod, true, pod.UID)
		service.onlineMutex.Unlock()
		log.Debugf("added online pod %v", pod.UID)
	case "offline":
		service.offlineMutex.Lock()
		offlinePod, err := newPodStatus(pod, false, pod.UID)
		if err != nil {
			service.offlineMutex.Unlock()
			return
		}
		service.offlineTasks[pod.UID] = offlinePod
		service.offlineMutex.Unlock()
		log.Debugf("added offline pod %v", pod.UID)
	}
}

func (service *CpiService) deletePod(pod *typedef.PodInfo) {
	switch pod.Annotations[constant.CpiAnnotationKey] {
	case "online":
		service.onlineMutex.Lock()
		delete(service.onlineTasks, pod.UID)
		service.onlineMutex.Unlock()
		log.Debugf("deleted online pod %v", pod.UID)
	case "offline":
		service.offlineMutex.Lock()
		delete(service.offlineTasks, pod.UID)
		service.offlineMutex.Unlock()
		log.Debugf("deleted offline pod %v", pod.UID)
	}
}

func (service *CpiService) getOnlinePodByUID(podUID string) (*podStatus, bool) {
	service.onlineMutex.RLock()
	onlinePodStatus, ok := service.onlineTasks[podUID]
	service.onlineMutex.RUnlock()
	return onlinePodStatus, ok
}

func (service *CpiService) getOfflinePodByUID(podUID string) (*podStatus, bool) {
	service.offlineMutex.RLock()
	offlinePodStatus, ok := service.offlineTasks[podUID]
	service.offlineMutex.RUnlock()
	return offlinePodStatus, ok
}

func (service *CpiService) getOnlinePods() map[string]*podStatus {
	onlinePods := make(map[string]*podStatus)
	service.onlineMutex.RLock()
	for UID, pod := range service.onlineTasks {
		onlinePods[UID] = pod
	}
	service.onlineMutex.RUnlock()
	return onlinePods
}

func (service *CpiService) getOfflinePods() map[string]*podStatus {
	offlinePods := make(map[string]*podStatus)
	service.offlineMutex.RLock()
	for UID, pod := range service.offlineTasks {
		offlinePods[UID] = pod
	}
	service.offlineMutex.RUnlock()
	return offlinePods
}

// newCpiService creates and initializes a new CpiService instance with the given name.
func newCpiService(name string) *CpiService {
	return &CpiService{
		ServiceBase: helper.ServiceBase{
			Name: name,
		},
		onlineTasks:  make(map[string]*podStatus, 0),
		offlineTasks: make(map[string]*podStatus, 0),
	}
}

// IsRunner returns true, indicating that the CPI service is persistent
func (service *CpiService) IsRunner() bool {
	return true
}

// Run starts the CPI service, which periodically collects data, identifies antagonists, and deletes expired data.
func (service *CpiService) Run(ctx context.Context) {
	go wait.Until(
		func() {
			service.collectData(defaultSampleDur)
		}, defaultCollectDur, ctx.Done())

	go wait.Until(func() {
		service.identifyAntagonists(defaultWindowDur, defaultIdentifyDur, defaultLimitDur, defaultAntagonistMetric, defaultLimitQuota)
	}, defaultIdentifyDur, ctx.Done())

	wait.Until(func() {
		service.deleteExpiredData(defaultWindowDur)
	}, defaultExpireDur, ctx.Done())
}

// PreStart initializes the CPI service by adding the current online and offline tasks into management.
func (service *CpiService) PreStart(viewer api.Viewer) error {
	if !perf.Support() {
		return fmt.Errorf("this machine does not support the Perf tool, so the CPI service cannot be executed. Please verify the system settings.")
	}
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	service.Viewer = viewer

	cpiPods := viewer.ListPodsWithOptions(func(pod *typedef.PodInfo) bool {
		_, ok := pod.Annotations[constant.CpiAnnotationKey]
		return ok
	})

	for _, cpiPod := range cpiPods {
		service.addPod(cpiPod)
	}
	return nil
}

// AddPod adds a pod to the CPI service's management, categorizing it as either online or offline
func (service *CpiService) AddPod(pod *typedef.PodInfo) error {
	_, ok := pod.Annotations[constant.CpiAnnotationKey]
	if !ok {
		return nil
	}
	service.addPod(pod)
	return nil
}

// DeletePod removes a pod from the CPI service's management.
func (service *CpiService) DeletePod(pod *typedef.PodInfo) error {
	_, ok := pod.Annotations[constant.CpiAnnotationKey]
	if !ok {
		return nil
	}
	service.deletePod(pod)
	return nil
}

// collectData gathers data for all online and offline tasks in the CPI service.
func (service *CpiService) collectData(sampleDur time.Duration) {
	now := time.Now()

	onlineTasks := service.getOnlinePods()
	for _, podStatus := range onlineTasks {
		go podStatus.collectData(sampleDur, now.UnixNano())
	}

	offlineTasks := service.getOfflinePods()
	for _, podStatus := range offlineTasks {
		go podStatus.collectData(sampleDur, now.UnixNano())
	}
}

// identifyAntagonists detects online pod outliers, identifies offline pods causing CPU interference, and limits their CPU quota.
func (service *CpiService) identifyAntagonists(windowDur, identifyDur, limitDur time.Duration, antagonistMetric float64, limitQuota string) {
	var outliers []string
	now := time.Now()
	onlineTasks := service.getOnlinePods()
	for podUID, podStatus := range onlineTasks {
		if podStatus.checkOutlier(now, identifyDur) {
			log.Debugf("identify outlier online pod: %v", podUID)
			outliers = append(outliers, podUID)
		}
	}
	if len(outliers) == 0 {
		return
	}

	var antagonists []string
	offlineTasks := service.getOfflinePods()
	for offlinePodUID, offlinePodStatus := range offlineTasks {
		for _, onlinePodUID := range outliers {
			onlinePodStatus, ok := service.getOnlinePodByUID(onlinePodUID)
			if !ok {
				continue
			}
			if onlinePodStatus.checkAntagonists(now, windowDur, offlinePodStatus, antagonistMetric) {
				log.Debugf("identified antagonist offline pod: %v", offlinePodUID)
				antagonists = append(antagonists, offlinePodUID)
				break
			}
		}
	}

	for _, podUID := range antagonists {
		podStatus, ok := service.getOfflinePodByUID(podUID)
		if !ok {
			continue
		}
		if err := podStatus.limit(limitQuota, limitDur); err != nil {
			log.Errorf("%v", err)
		}
	}
	if len(antagonists) == 0 && len(service.offlineTasks) != 0 {
		service.limitMaxUsageOffline(windowDur, limitQuota, limitDur)
	}
}

func (service *CpiService) limitMaxUsageOffline(windowDur time.Duration, limitQuota string, limitDur time.Duration) {
	var maxUsage float64
	var maxUsagePod *podStatus
	now := time.Now()
	offlineTasks := service.getOfflinePods()
	for _, podStatus := range offlineTasks {
		cpuUsages := podStatus.rangeSearch(CPUUsage, now.Add(-windowDur).UnixNano(), now.UnixNano()) //rangeSearch
		var usageSum float64
		for _, cpuUsage := range cpuUsages.data {
			usageSum += cpuUsage
		}
		if usageSum > maxUsage {
			maxUsagePod = podStatus
			maxUsage = usageSum
		}
	}
	if maxUsagePod == nil {
		return
	}
	if err := maxUsagePod.limit(limitQuota, limitDur); err != nil {
		log.Errorf("%v", err)
		return
	}
}

// deleteExpiredData removes expired data from both online and offline tasks based on the specified expiration duration.
func (service *CpiService) deleteExpiredData(expireDur time.Duration) {
	now := time.Now()
	onlineTasks := service.getOnlinePods()
	for _, onlinePodStatus := range onlineTasks {
		onlinePodStatus.expireAndUpdateThreshold(now.Add(-1 * expireDur).UnixNano())
	}

	offlineTasks := service.getOfflinePods()
	for _, offlinePodStatus := range offlineTasks {
		offlinePodStatus.expireAndUpdateThreshold(now.Add(-1 * expireDur).UnixNano())
	}
}

func (service *CpiService) Terminate(api.Viewer) error {
	offlineTasks := service.getOfflinePods()
	for _, pods := range offlineTasks {
		//If recoverQuota encounters an error, it will not attempt to recover again; instead, it will output the error itself, so the retry interval is set to 1 hour."
		pods.recoverQuota(time.Hour)
	}
	return nil
}
