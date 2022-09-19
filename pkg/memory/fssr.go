// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2022-9-2
// Description:
// 1. When Rubik starts, all offline memory.high is configured to 80% of total memory by default.
// 2. When memory pressure increases: Available memory freeMemory < reservedMemory(totalMemory * 5%).
// newly memory.high=memory.high-totalMemory * 10%.
// 3. When memory is rich over a period of time: freeMemory > 3 * reservedMemory, In this case, 1% of
// the totalMemory is reserved for offline applications. High=memory.high+totalMemory * 1% until memory
// free is between reservedMemory and 3 * reservedMemory.

// Package memory provide memory reclaim strategy for offline tasks.
package memory

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
)

type fssrStatus int

const (
	reservePercentage   = 0.05
	waterlinePercentage = 0.8
	relievePercentage   = 0.02
	reclaimPercentage   = 0.1
	prerelieveInterval  = "30m"
	reserveRatio        = 3
	highAsyncRatio      = 90
)

const (
	fssrNormal fssrStatus = iota
	fssrReclaim
	fssrPreRelieve
	fssrRelieve
)

type fssr struct {
	mmgr                *MemoryManager
	preRelieveStartDate time.Time
	st                  fssrStatus
	total               int64
	limit               int64
	reservedMemory      int64
	highAsyncRatio      int64
}

func newFssr(m *MemoryManager) (f *fssr) {
	f = new(fssr)
	f.init(m)
	return f
}

func (f *fssr) init(m *MemoryManager) {
	memInfo, err := getMemoryInfo()
	if err != nil {
		log.Infof("initialization of fssr failed")
		return
	}

	f.mmgr = m
	f.total = memInfo.total
	f.reservedMemory = int64(reservePercentage * float64(f.total))
	f.limit = int64(waterlinePercentage * float64(f.total))
	f.st = fssrNormal
	f.highAsyncRatio = highAsyncRatio
	f.initOfflineContainerLimit()

	log.Infof("total: %v, reserved Memory: %v, limit memory: %v", f.total, f.reservedMemory, f.limit)
}

func (f *fssr) Run() {
	go wait.Until(f.timerProc, time.Duration(f.mmgr.checkInterval)*time.Second, f.mmgr.stop)
}

// UpdateConfig is used to update memory config
func (f *fssr) UpdateConfig(pod *typedef.PodInfo) {
	for _, c := range pod.Containers {
		f.initContainerMemoryLimit(c)
	}
}

func (f *fssr) timerProc() {
	f.updateStatus()
	if f.needAdjust() {
		newLimit := f.calculateNewLimit()
		f.adjustOfflineContainerMemory(newLimit)
	}
}

func (f *fssr) initOfflineContainerLimit() {
	if f.mmgr.cpm == nil {
		log.Infof("init offline container limit failed, cpm is nil")
		return
	}

	containers := f.mmgr.cpm.ListOfflineContainers()
	for _, c := range containers {
		f.initContainerMemoryLimit(c)
	}
}

func (f *fssr) needAdjust() bool {
	if f.st == fssrReclaim || f.st == fssrRelieve {
		return true
	}
	return false
}

func (f *fssr) updateStatus() {
	curMemInfo, err := getMemoryInfo()
	if err != nil {
		log.Errorf("get memory info failed, err:%v", err)
		return
	}
	oldStatus := f.st

	// Use free instead of Available
	if curMemInfo.free < f.reservedMemory {
		f.st = fssrReclaim
	} else if curMemInfo.free > reserveRatio*f.reservedMemory {
		switch f.st {
		case fssrNormal:
			f.st = fssrPreRelieve
			f.preRelieveStartDate = time.Now()
		case fssrPreRelieve, fssrRelieve:
			t, _ := time.ParseDuration(prerelieveInterval)
			if f.preRelieveStartDate.Add(t).Before(time.Now()) {
				f.st = fssrRelieve
			}
		case fssrReclaim:
			f.st = fssrNormal
		default:
			log.Errorf("status incorrect, this should not happen")
		}
	}

	log.Infof("update change status from %v to %v, cur available %v, cur free %v",
		oldStatus, f.st, curMemInfo.available, curMemInfo.free)
}

func (f *fssr) calculateNewLimit() int64 {
	newLimit := f.limit
	if f.st == fssrReclaim {
		newLimit = f.limit - int64(reclaimPercentage*float64(f.total))
		if newLimit < 0 || newLimit <= f.reservedMemory {
			newLimit = f.reservedMemory
			log.Infof("reclaim offline containers current limit %v is too small, set as reserved memory %v", newLimit, f.reservedMemory)
		}
	} else if f.st == fssrRelieve {
		newLimit = f.limit + int64(relievePercentage*float64(f.total))
		if newLimit > int64(waterlinePercentage*float64(f.total)) {
			newLimit = int64(waterlinePercentage * float64(f.total))
			log.Infof("relieve offline containers limit soft memory exceeds waterline, set limit as waterline %v", waterlinePercentage*float64(f.total))
		}
	}
	return newLimit
}

func (f *fssr) initContainerMemoryLimit(c *typedef.ContainerInfo) {
	path := c.CgroupPath("memory")
	if err := writeMemoryLimit(path, typedef.FormatInt64(f.limit), mhigh); err != nil {
		log.Errorf("failed to initialize the limit soft memory of offline container %v: %v", c.ID, err)
	} else {
		log.Infof("initialize the limit soft memory of the offline container %v to %v successfully", c.ID, f.limit)
	}

	if err := writeMemoryLimit(path, typedef.FormatInt64(f.highAsyncRatio), mhighAsyncRatio); err != nil {
		log.Errorf("failed to initialize the async high ration of offline container %v: %v", c.ID, err)
	} else {
		log.Infof("initialize the async high ration of the offline container %v:%v success", c.ID, f.highAsyncRatio)
	}
}

func (f *fssr) adjustOfflineContainerMemory(limit int64) {
	if f.mmgr.cpm == nil {
		log.Infof("reclaim offline containers failed, cpm is nil")
		return
	}

	containers := f.mmgr.cpm.ListOfflineContainers()
	for _, c := range containers {
		path := c.CgroupPath("memory")
		if err := writeMemoryLimit(path, typedef.FormatInt64(limit), mhigh); err != nil {
			log.Errorf("relieve offline containers limit soft memory %v failed, err is %v", c.ID, err)
		} else {
			f.limit = limit
		}
	}
}
