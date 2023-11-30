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
// Description: QuotaTurbo Status Store

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// cpuUtil is used to store the cpu usage at a specific time
type cpuUtil struct {
	timestamp int64
	util      float64
}

// StatusStore is the information of node/containers obtained for quotaTurbo
type StatusStore struct {
	// configuration of the QuotaTurbo
	*Config
	// ensuring Concurrent Sequential Consistency
	sync.RWMutex
	// map between container IDs and container CPU quota
	cpuQuotas map[string]*CPUQuota
	// cpu utilization sequence for N consecutive cycles
	cpuUtils []cpuUtil
	// /proc/stat of the previous period
	lastProcStat ProcStat
}

// NewStatusStore returns a pointer to StatusStore
func NewStatusStore() *StatusStore {
	return &StatusStore{
		Config: NewConfig(),
		lastProcStat: ProcStat{
			total: -1,
			busy:  -1,
		},
		cpuQuotas: make(map[string]*CPUQuota, 0),
		cpuUtils:  make([]cpuUtil, 0),
	}
}

// AddCgroup adds cgroup need to be adjusted
func (store *StatusStore) AddCgroup(cgroupPath string, cpuLimit float64) error {
	if len(cgroupPath) == 0 {
		return fmt.Errorf("cgroup path should not be empty")
	}
	if store.CgroupRoot == "" {
		return fmt.Errorf("undefined cgroup mount point, please set it firstly")
	}
	h := cgroup.NewHierarchy(store.CgroupRoot, cgroupPath)
	if !isAdjustmentAllowed(h, cpuLimit) {
		return fmt.Errorf("cgroup not allow to adjust")
	}
	c, err := NewCPUQuota(h, cpuLimit)
	if err != nil {
		return fmt.Errorf("error creating cpu quota: %v", err)
	}
	store.Lock()
	store.cpuQuotas[cgroupPath] = c
	store.Unlock()
	return nil
}

// RemoveCgroup deletes cgroup that do not need to be adjusted.
func (store *StatusStore) RemoveCgroup(cgroupPath string) error {
	store.RLock()
	cq, ok := store.cpuQuotas[cgroupPath]
	store.RUnlock()
	if !ok {
		return nil
	}
	safeDel := func(id string) error {
		store.Lock()
		delete(store.cpuQuotas, id)
		store.Unlock()
		return nil
	}

	if !util.PathExist(filepath.Join(cq.MountPoint, "cpu", cq.Path)) {
		return safeDel(cgroupPath)
	}
	if err := cq.recoverQuota(); err != nil {
		return fmt.Errorf("failed to recover cpu.cfs_quota_us for cgroup %s : %v", cq.Path, err)
	}
	return safeDel(cgroupPath)
}

// AllCgroups returns all cgroup paths that are adjusting quota
func (store *StatusStore) AllCgroups() []string {
	var res = make([]string, 0)
	for _, cq := range store.cpuQuotas {
		res = append(res, cq.Path)
	}
	return res
}

// getLastCPUUtil obtain the latest cpu utilization
func (store *StatusStore) getLastCPUUtil() float64 {
	if len(store.cpuUtils) == 0 {
		return 0
	}
	return store.cpuUtils[len(store.cpuUtils)-1].util
}

// updateCPUUtils updates the cpu usage of a node
func (store *StatusStore) updateCPUUtils() error {
	var (
		curUtil float64
		index   int
		t       cpuUtil
	)
	ps, err := getProcStat()
	if err != nil {
		return err
	}
	if store.lastProcStat.total >= 0 {
		curUtil = calculateUtils(store.lastProcStat, ps)
	}
	store.lastProcStat = ps
	cur := time.Now().UnixNano()
	store.cpuUtils = append(store.cpuUtils, cpuUtil{
		timestamp: cur,
		util:      curUtil,
	})
	// retain utilization data for only one minute
	const minuteTimeDelta = int64(time.Minute)
	for index, t = range store.cpuUtils {
		if cur-t.timestamp <= minuteTimeDelta {
			break
		}
	}
	if index > 0 {
		store.cpuUtils = store.cpuUtils[index:]
	}
	return nil
}

func (store *StatusStore) updateCPUQuotas() error {
	var errs error
	for id, c := range store.cpuQuotas {
		if err := c.update(); err != nil {
			errs = appendErr(errs, fmt.Errorf("failed to update cpu quota %v: %v", id, err))
		}
	}
	return errs
}

// writeQuota writes the calculated quota value into the cgroup file and takes effect
func (store *StatusStore) writeQuota() error {
	var errs error
	for id, c := range store.cpuQuotas {
		if err := c.writeQuota(); err != nil {
			errs = appendErr(errs, fmt.Errorf("failed to write cgroup quota %v: %v", id, err))
		}
	}
	return errs
}

// isAdjustmentAllowed judges whether quota adjustment is allowed
func isAdjustmentAllowed(h *cgroup.Hierarchy, cpuLimit float64) bool {
	// 1. containers whose cgroup path does not exist are not considered.
	if !util.PathExist(filepath.Join(h.MountPoint, "cpu", h.Path)) {
		return false
	}

	/*
		2. abnormal CPULimit
		 a). containers that do not limit the quota => cpuLimit = 0
		 b). cpuLimit = 0 : k8s allows the CPULimit to be 0, but the quota is not limited.
		 c). cpuLimit >= all cores
	*/
	if cpuLimit <= 0 ||
		cpuLimit >= float64(runtime.NumCPU()) {
		return false
	}
	return true
}
