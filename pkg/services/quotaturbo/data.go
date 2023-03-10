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
// Date: 2023-02-20
// Description: QuotaTurbo driver interface

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	defaultHightWaterMark         = 60
	defaultAlarmWaterMark         = 80
	defaultQuotaTurboSyncInterval = 100
)

// cpuUtil is used to store the cpu usage at a specific time
type cpuUtil struct {
	timestamp int64
	util      float64
}

// Config defines configuration of QuotaTurbo
type Config struct {
	HighWaterMark  int `json:"highWaterMark,omitempty"`
	AlarmWaterMark int `json:"alarmWaterMark,omitempty"`
	SyncInterval   int `json:"syncInterval,omitempty"`
}

// NodeData is the information of node/containers obtained for quotaTurbo
type NodeData struct {
	// configuration of the QuotaTurbo
	*Config
	// ensuring Concurrent Sequential Consistency
	sync.RWMutex
	// map between container IDs and container CPU quota
	containers map[string]*CPUQuota
	// cpu utilization sequence for N consecutive cycles
	cpuUtils []cpuUtil
	// /proc/stat of the previous period
	lastProcStat ProcStat
}

// NewNodeData returns a pointer to NodeData
func NewNodeData() *NodeData {
	return &NodeData{
		Config: &Config{
			HighWaterMark:  defaultHightWaterMark,
			AlarmWaterMark: defaultAlarmWaterMark,
			SyncInterval:   defaultQuotaTurboSyncInterval,
		},
		lastProcStat: ProcStat{
			total: -1,
			busy:  -1,
		},
		containers: make(map[string]*CPUQuota, 0),
		cpuUtils:   make([]cpuUtil, 0),
	}
}

// getLastCPUUtil obtain the latest cpu utilization
func (d *NodeData) getLastCPUUtil() float64 {
	if len(d.cpuUtils) == 0 {
		return 0
	}
	return d.cpuUtils[len(d.cpuUtils)-1].util
}

// removeContainer deletes the list of pods that do not need to be adjusted.
func (d *NodeData) removeContainer(id string) error {
	d.RLock()
	cq, ok := d.containers[id]
	d.RUnlock()
	if !ok {
		return nil
	}
	safeDel := func(id string) error {
		d.Lock()
		delete(d.containers, id)
		d.Unlock()
		return nil
	}

	if !util.PathExist(cgroup.AbsoluteCgroupPath("cpu", cq.CgroupPath)) {
		return safeDel(id)
	}
	// cq.Period ranges from 1000(us) to 1000000(us) and does not overflow.
	origin := int64(cq.LimitResources[typedef.ResourceCPU] * float64(cq.period))
	if err := cq.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(origin)); err != nil {
		return fmt.Errorf("fail to recover cpu.cfs_quota_us for container %s : %v", cq.Name, err)
	}
	return safeDel(id)
}

// updateCPUUtils updates the cpu usage of a node
func (d *NodeData) updateCPUUtils() error {
	var (
		curUtil float64 = 0
		index           = 0
		t       cpuUtil
	)
	ps, err := getProcStat()
	if err != nil {
		return err
	}
	if d.lastProcStat.total >= 0 {
		curUtil = calculateUtils(d.lastProcStat, ps)
	}
	d.lastProcStat = ps
	cur := time.Now().UnixNano()
	d.cpuUtils = append(d.cpuUtils, cpuUtil{
		timestamp: cur,
		util:      curUtil,
	})
	// retain utilization data for only one minute
	const minuteTimeDelta = int64(time.Minute)
	for index, t = range d.cpuUtils {
		if cur-t.timestamp <= minuteTimeDelta {
			break
		}
	}
	if index > 0 {
		d.cpuUtils = d.cpuUtils[index:]
	}
	return nil
}

// UpdateClusterContainers synchronizes data from given containers
func (d *NodeData) UpdateClusterContainers(conts map[string]*typedef.ContainerInfo) error {
	var toBeDeletedList []string
	for _, cont := range conts {
		old, ok := d.containers[cont.ID]
		// delete or skip containers that do not meet the conditions.
		if !isAdjustmentAllowed(cont) {
			if ok {
				toBeDeletedList = append(toBeDeletedList, cont.ID)
			}
			continue
		}
		// add container
		if !ok {
			log.Debugf("add container %v (name : %v)", cont.ID, cont.Name)
			if newQuota, err := NewCPUQuota(cont); err != nil {
				log.Errorf("failed to create cpu quota object %v, error: %v", cont.Name, err)
			} else {
				d.containers[cont.ID] = newQuota
			}
			continue
		}
		// update data container in the quotaTurboList
		old.ContainerInfo = cont
		if err := old.updatePeriod(); err != nil {
			log.Errorf("fail to update period : %v", err)
		}
		if err := old.updateThrottle(); err != nil {
			log.Errorf("fail to update throttle time : %v", err)
		}
		if err := old.updateQuota(); err != nil {
			log.Errorf("fail to update quota : %v", err)
		}
		if err := old.updateUsage(); err != nil {
			log.Errorf("fail to update cpu usage : %v", err)
		}
	}
	// non trust list container
	for id := range d.containers {
		// the container is removed from the trust list
		if _, ok := conts[id]; !ok {
			toBeDeletedList = append(toBeDeletedList, id)
		}
	}

	for _, id := range toBeDeletedList {
		if err := d.removeContainer(id); err != nil {
			log.Errorf(err.Error())
		}
	}
	return nil
}

// WriteQuota saves the quota value of the container
func (d *NodeData) WriteQuota() {
	for _, c := range d.containers {
		if err := c.WriteQuota(); err != nil {
			log.Errorf(err.Error())
		}
	}
}

// isAdjustmentAllowed judges whether quota adjustment is allowed
func isAdjustmentAllowed(ci *typedef.ContainerInfo) bool {
	// 1. containers whose cgroup path does not exist are not considered.
	if !util.PathExist(cgroup.AbsoluteCgroupPath("cpu", ci.CgroupPath, "")) {
		return false
	}

	// 2. abnormal CPULimit
	// a). containers that do not limit the quota
	// b). CPULimit = 0 : k8s allows the CPULimit to be 0, but the quota is not limited.
	if ci.LimitResources[typedef.ResourceCPU] <= 0 ||
		ci.RequestResources[typedef.ResourceCPU] <= 0 ||
		ci.LimitResources[typedef.ResourceCPU] == float64(runtime.NumCPU()) ||
		ci.RequestResources[typedef.ResourceCPU] == float64(runtime.NumCPU()) {
		return false
	}
	return true
}
