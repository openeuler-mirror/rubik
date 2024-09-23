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
// Description: This file provides methods and class for managing Pod status.

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"fmt"
	"math"
	"sync"
	"time"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

var cpiConf []int = []int{perf.INSTRUCTIONS, perf.CYCLES, perf.CACHEREFERENCES, perf.CACHEMISS}

const (
	CPI = iota
	CPUUsage
)

type containerStatus struct {
	UID         string
	preCpuQuota string
	*cgroup.Hierarchy
}
type podStatus struct {
	UID        string
	isOnline   bool
	containers map[string]*containerStatus
	*cgroup.Hierarchy
	cpiSeries      *dataSeries
	cpuUsageSeries *dataSeries
	cpiMean        float64
	stdDev         float64
	count          int64
	isLimited      bool
	preCpuQuota    string
	delayer        *time.Timer
	podMutex       sync.RWMutex
}

func (podStatus *podStatus) setCPIStatistic(count int64, cpiMean float64, stdDev float64) {
	podStatus.podMutex.Lock()
	podStatus.count = count
	podStatus.cpiMean = cpiMean
	podStatus.stdDev = stdDev
	podStatus.podMutex.Unlock()
}

func (podStatus *podStatus) getCPIStatistic() (int64, float64, float64) {
	podStatus.podMutex.RLock()
	count := podStatus.count
	cpiMean := podStatus.cpiMean
	stdDev := podStatus.stdDev
	podStatus.podMutex.RUnlock()
	return count, cpiMean, stdDev
}

func (podStatus *podStatus) addData(metric int, data float64, nowNano int64) error {
	var err error
	podStatus.podMutex.Lock()
	switch metric {
	case CPI:
		err = podStatus.cpiSeries.add(data, nowNano)
	case CPUUsage:
		err = podStatus.cpuUsageSeries.add(data, nowNano)
	}
	podStatus.podMutex.Unlock()
	return err
}

// cpuUsageExpiredSeries := podStatus.cpuUsageSeries.expire(expireNano)
func (podStatus *podStatus) expire(expireNano int64) (*dataSeries, *dataSeries) {
	podStatus.podMutex.Lock()
	expiredCpuUsage := podStatus.cpuUsageSeries.expire(expireNano)
	if !podStatus.isOnline {
		podStatus.podMutex.Unlock()
		return nil, expiredCpuUsage
	}
	expiredCpi := podStatus.cpiSeries.expire(expireNano)
	podStatus.podMutex.Unlock()
	return expiredCpi, expiredCpuUsage
}

func (podStatus *podStatus) rangeSearch(metric int, startTimeNano int64, endTimeNano int64) *dataSeries {
	var rangeData *dataSeries
	podStatus.podMutex.RLock()
	switch metric {
	case CPI:
		rangeData = podStatus.cpiSeries.rangeSearch(int64(startTimeNano), int64(endTimeNano))
	case CPUUsage:
		rangeData = podStatus.cpuUsageSeries.rangeSearch(int64(startTimeNano), int64(endTimeNano))
	}
	podStatus.podMutex.RUnlock()
	return rangeData
}

// containerStatus.GetCgroupAttr(cpuQuotaKey)
func (podStatus *podStatus) getAndSetOriginQuota() error {
	quotaAttr := podStatus.GetCgroupAttr(cpuQuotaKey)
	if quotaAttr.Err != nil {
		return fmt.Errorf("failed to get pod %v CPU quota: %v", podStatus.UID, quotaAttr.Err)
	}
	podStatus.preCpuQuota = quotaAttr.Value

	if podStatus.isOnline {
		return nil
	}

	for _, containerStatus := range podStatus.containers {
		quotaAttr := containerStatus.GetCgroupAttr(cpuQuotaKey)
		if quotaAttr.Err != nil {
			return fmt.Errorf("failed to get containerd %v CPU quota: %v", containerStatus.UID, quotaAttr.Err)
		}
		containerStatus.preCpuQuota = quotaAttr.Value
	}
	return nil
}

func (podStatus *podStatus) limitQuota(limitedQuota string) error {
	podStatus.podMutex.Lock()
	limietedHierachy := make(map[*cgroup.Hierarchy]string)
	for _, containerStatus := range podStatus.containers {
		err := containerStatus.SetCgroupAttr(cpuQuotaKey, limitedQuota)
		if err == nil {
			limietedHierachy[containerStatus.Hierarchy] = containerStatus.preCpuQuota
			continue
		}
		for hierachy, prequota := range limietedHierachy {
			if err := hierachy.SetCgroupAttr(cpuQuotaKey, prequota); err != nil {
				return fmt.Errorf("recover CPU quota %v failed when limite container %v CPU  quota failed: %v", hierachy.Path, containerStatus.UID, err)
			}
		}
		return fmt.Errorf("limit container %v CPU quota failed: %v", containerStatus.UID, err)
	}
	if err := podStatus.SetCgroupAttr(cpuQuotaKey, limitedQuota); err != nil {
		for hierachy, prequota := range limietedHierachy {
			if err := hierachy.SetCgroupAttr(cpuQuotaKey, prequota); err != nil {
				return fmt.Errorf("recover CPU quota %v failed when limite pod %v CPU quota failed: %v", hierachy.Path, podStatus.UID, err)
			}
		}
		return fmt.Errorf("limit pod %v CPU quota failed: %v", podStatus.UID, err)
	}
	podStatus.isLimited = true
	podStatus.podMutex.Unlock()
	log.Debugf("offlinePod %v has been limited", podStatus.UID)
	return nil
}

func (podStatus *podStatus) recoverQuota(limitDur time.Duration) {
	podStatus.podMutex.Lock()
	if err := podStatus.SetCgroupAttr(cpuQuotaKey, podStatus.preCpuQuota); err != nil {
		log.Errorf("failed to recover pod %v CPU quota: %v", podStatus.UID, err)
		podStatus.delayer.Reset(limitDur)
		return
	}
	for _, containerStatus := range podStatus.containers {
		if err := containerStatus.SetCgroupAttr(cpuQuotaKey, containerStatus.preCpuQuota); err != nil {
			log.Errorf("failed to recover container %v CPU quota: %v", containerStatus.UID, err)
			podStatus.delayer.Reset(limitDur)
			return
		}
	}
	podStatus.isLimited = false
	podStatus.podMutex.Unlock()
	log.Debugf("offlinePod %v has been recovered", podStatus.UID)
}

// newPodStatus creates and returns a new podStatus based on the pod's online status.
func newPodStatus(pod *typedef.PodInfo, isOnline bool, uid string) (*podStatus, error) {
	if isOnline {
		return &podStatus{
			UID:            uid,
			isOnline:       isOnline,
			Hierarchy:      &pod.Hierarchy,
			cpiSeries:      newDataSeries(),
			cpuUsageSeries: newDataSeries(),
		}, nil
	}
	containers := make(map[string]*containerStatus)
	for containerUID, containerInfo := range pod.IDContainersMap {
		containers[containerUID] = newContainerStatus(&containerInfo.Hierarchy, containerUID)
	}
	offlinePod := &podStatus{
		UID:            uid,
		isOnline:       isOnline,
		Hierarchy:      &pod.Hierarchy,
		cpuUsageSeries: newDataSeries(),
		containers:     containers,
	}
	if err := offlinePod.getAndSetOriginQuota(); err != nil {
		return offlinePod, err
	}
	return offlinePod, nil
}

// newContainerStatus creates and returns a new containerStatus with the specified cgroup hierarchy
func newContainerStatus(h *cgroup.Hierarchy, uid string) *containerStatus {
	return &containerStatus{
		UID:       uid,
		Hierarchy: h,
	}
}

// collectData collects and records CPU usage and CPI data for the pod during the given time duration.
func (podStatus *podStatus) collectData(sampleDur time.Duration, nowNano int64) {
	if podStatus.isOnline {
		go podStatus.collectCPI(sampleDur, nowNano)
	}
	podStatus.collectCPUUsage(sampleDur, nowNano)
}

func (podStatus *podStatus) collectCPI(sampleDur time.Duration, nowNano int64) {
	cgroupPath := cgroup.AbsoluteCgroupPath("perf_event", podStatus.Path, "")
	if !util.PathExist(cgroupPath) {
		log.Errorf("cgroup path does not exist: %v", cgroupPath)
		return
	}
	stat, err := perf.CgroupStat(cgroupPath, sampleDur, cpiConf)
	if err != nil {
		log.Errorf("failed to collect perf data from cgroup: %v", err)
		return
	}
	if stat.Instructions == 0 {
		return
	}

	cpi := (float64)(stat.CPUCycles) / (float64)(stat.Instructions)
	log.Debugf("onlinePod collected cpi = %v", cpi)

	if err := podStatus.addData(CPI, cpi, nowNano); err != nil {
		log.Errorf("collectData invoke add: %v", err)
	}
}

func (podStatus *podStatus) collectCPUUsage(sampleDur time.Duration, nowNano int64) {
	startUsage, err := podStatus.GetCgroupAttr(cpuUsageKey).Int64()
	if err != nil {
		log.Errorf("failed to get CPU Usage: %v", err)
		return
	}

	time.Sleep(sampleDur)

	endUsage, err := podStatus.GetCgroupAttr(cpuUsageKey).Int64()
	if err != nil {
		log.Errorf("failed to get CPU Usage after %v: %v", sampleDur, err) //在podStatus中加上pod UID
		return
	}

	if err := podStatus.addData(CPUUsage, float64(endUsage-startUsage)/1e9/float64(sampleDur.Seconds()), nowNano); err != nil {
		log.Errorf("collectData invoke add: %v", err)
	}
}

// expireAndUpdateThreshold removes expired data and updates the CPI threshold for the pod.
func (podStatus *podStatus) expireAndUpdateThreshold(expireNano int64) {
	expiredCpiSeries, expiredCpuSeries := podStatus.expire(expireNano)
	if !podStatus.isOnline {
		return
	}
	count, cpiMean, stdDev := podStatus.getCPIStatistic()
	//If CPU usage is less than 0.25 or CPI is greater than the threshold, it indicates low metric availability and should not be used to update CPI statistic
	for nano, cpi := range expiredCpiSeries.data {
		cpuUsage, ok := expiredCpuSeries.data[nano]
		if !ok {
			continue
		}
		//When CPU usage is less than 0.25, CPI is not accurate and should not be used to update CPI statistic.
		if cpuUsage < 0.25 {
			delete(expiredCpiSeries.data, nano)
		} else if count >= defaultMinCount && cpi > cpiMean+2*stdDev {
			delete(expiredCpiSeries.data, nano)
		}
	}
	podStatus.updateCPIStatistic(expiredCpiSeries)
}

// updateThreshold recalculates and updates the CPI mean and standard deviation based on expired data.
func (podStatus *podStatus) updateCPIStatistic(expired *dataSeries) {
	if len(expired.data) == 0 {
		return
	}

	var expiredCpiSum float64 = 0
	for _, value := range expired.data {
		expiredCpiSum += value
	}

	expiredMean := expiredCpiSum / float64(len(expired.data))
	var expiredVar float64
	for _, value := range expired.data {
		expiredVar += (value - expiredMean) * (value - expiredMean)
	}

	count, cpiMean, stdDev := podStatus.getCPIStatistic()

	newMean := ((float64(count) * cpiMean) + expiredCpiSum) / float64(count+int64(len(expired.data)))
	originVar := stdDev * stdDev
	combinedVar := (float64(count)*(originVar+(cpiMean-newMean)*(cpiMean-newMean)) +
		float64(len(expired.data))*(expiredVar/float64(len(expired.data))+(expiredMean-newMean)*(expiredMean-newMean))) / float64(count+int64(len(expired.data)))

	if combinedVar < 0 {
		combinedVar = 0
	}
	newStdDev := math.Sqrt(combinedVar)
	newCount := count + int64(len(expired.data))
	log.Debugf("new Mean = %v , new StdDev = %v", newMean, newStdDev)
	podStatus.setCPIStatistic(newCount, newMean, newStdDev)
}

// checkOutlier checks if the pod has at least three CPI is bigger than CPI threshold within the specified time duration.
func (podStatus *podStatus) checkOutlier(now time.Time, duration time.Duration) bool {
	count, cpiMean, stdDev := podStatus.getCPIStatistic()
	// count <= defaultMinCount indicates that the sample count used to calculate the CPI statistic is too small and cannot be used for interference detection.
	if count <= defaultMinCount {
		return false
	}

	threshold := cpiMean + 2*stdDev
	cpiValues := podStatus.rangeSearch(CPI, now.Add(-duration).UnixNano(), now.UnixNano())
	cpuSeries := podStatus.rangeSearch(CPUUsage, now.Add(-duration).UnixNano(), now.UnixNano())
	if cpiValues == nil {
		return false
	}
	outlieCount := 0

	//When CPU usage is less than 0.25, CPI is not accurate and should not be used to determine whether interference has occurred.
	for t, cpi := range cpiValues.data {
		cpuUsage, ok := cpuSeries.getData(t)
		if !ok || cpuUsage < 0.25 || cpi <= threshold {
			continue
		}
		outlieCount++
	}
	return outlieCount >= 3
}

// checkAntagonists checks if there is a correlation between the pod's CPI and an offline pod's CPU usage that exceeds the given antagonist metric.
func (podStatus *podStatus) checkAntagonists(now time.Time, window time.Duration, offlinePod *podStatus, antagonistMetric float64) bool {
	windowStart := now.Add(-1 * window)
	_, cpiMean, stdDev := podStatus.getCPIStatistic()

	victimCpi := podStatus.rangeSearch(CPI, windowStart.UnixNano(), now.UnixNano())
	suspectCpu := offlinePod.rangeSearch(CPUUsage, windowStart.UnixNano(), now.UnixNano()).normalize()

	if len(victimCpi.timeline) == 0 || len(suspectCpu.timeline) == 0 {
		return false
	}

	threshold := cpiMean + 2*stdDev
	var correlation float64
	for t, cpi := range victimCpi.data {
		cpuUsage, ok := suspectCpu.getData(t)
		if !ok {
			continue
		}
		if cpi > threshold {
			correlation += cpuUsage * (1 - threshold/cpi)
		} else {
			correlation += cpuUsage * (cpi/threshold - 1)
		}
	}
	return correlation > antagonistMetric
}

// limitCpuQuota limits the CPU quota for the pod and its containers for a specified duration and resets after that period.
func (podStatus *podStatus) limit(quota string, limitDur time.Duration) error {
	if podStatus.isLimited {
		podStatus.delayer.Reset(limitDur)
		return nil
	}

	if err := podStatus.limitQuota(quota); err != nil {
		return err
	}
	if podStatus.delayer != nil {
		podStatus.delayer.Reset(limitDur)
		return nil
	}

	recoverQuota := func() {
		podStatus.recoverQuota(limitDur)
	}
	podStatus.delayer = time.AfterFunc(limitDur, recoverQuota)
	return nil
}
