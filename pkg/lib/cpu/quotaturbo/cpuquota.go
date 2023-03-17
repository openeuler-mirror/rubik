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
// Description: cpu container cpu quota data and methods

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
	"path"
	"time"

	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	// numberOfRestrictedCycles is the number of periods in which the quota limits the CPU usage.
	numberOfRestrictedCycles = 60
	// The default value of the cfs_period_us file is 100ms
	defaultCFSPeriodUs int64 = 100000
)

var (
	cpuPeriodKey    = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_period_us"}
	cpuQuotaKey     = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_quota_us"}
	cpuAcctUsageKey = &cgroup.Key{SubSys: "cpuacct", FileName: "cpuacct.usage"}
	cpuStatKey      = &cgroup.Key{SubSys: "cpu", FileName: "cpu.stat"}
)

// cpuUsage cpu time used by the container at timestamp
type cpuUsage struct {
	timestamp int64
	usage     int64
}

// CPUQuota stores the CPU quota information of a single container.
type CPUQuota struct {
	// basic container information
	*cgroup.Hierarchy
	// expect cpu limit
	cpuLimit float64
	// current throttling data for the container
	curThrottle *cgroup.CPUStat
	// previous throttling data for container
	preThrottle *cgroup.CPUStat
	// container cfs_period_us
	period int64
	// current cpu quota of the container
	curQuota int64
	// cpu quota of the container in the next period
	nextQuota int64
	// the delta of the cpu quota to be adjusted based on the decision.
	quotaDelta float64
	// the upper limit of the container cpu quota
	heightLimit float64
	// maximum quota that can be used by a container in the next period,
	// calculated based on the total usage in the past N-1 cycles
	maxQuotaNextPeriod float64
	// container cpu usage sequence
	cpuUsages []cpuUsage
}

// NewCPUQuota create a cpu quota object
func NewCPUQuota(h *cgroup.Hierarchy, cpuLimit float64) (*CPUQuota, error) {
	var defaultQuota = cpuLimit * float64(defaultCFSPeriodUs)
	cq := &CPUQuota{
		Hierarchy:          h,
		cpuLimit:           cpuLimit,
		cpuUsages:          make([]cpuUsage, 0),
		quotaDelta:         0,
		curThrottle:        &cgroup.CPUStat{NrThrottled: 0, ThrottledTime: 0},
		preThrottle:        &cgroup.CPUStat{NrThrottled: 0, ThrottledTime: 0},
		period:             defaultCFSPeriodUs,
		curQuota:           int64(defaultQuota),
		nextQuota:          int64(defaultQuota),
		heightLimit:        defaultQuota,
		maxQuotaNextPeriod: defaultQuota,
	}
	if err := cq.update(); err != nil {
		return cq, err
	}
	// The throttle data before and after the initialization is the same.
	cq.preThrottle = cq.curThrottle
	return cq, nil
}

func (c *CPUQuota) update() error {
	var errs error
	if err := c.updatePeriod(); err != nil {
		errs = appendErr(errs, err)
	}
	if err := c.updateThrottle(); err != nil {
		errs = appendErr(errs, err)
	}
	if err := c.updateQuota(); err != nil {
		errs = appendErr(errs, err)
	}
	if err := c.updateUsage(); err != nil {
		errs = appendErr(errs, err)
	}
	if errs != nil {
		return errs
	}
	return nil
}

func (c *CPUQuota) updatePeriod() error {
	us, err := c.GetCgroupAttr(cpuPeriodKey).Int64()
	// If an error occurs, the period remains unchanged or the default value is used.
	if err != nil {
		return err
	}
	c.period = us
	return nil
}

func (c *CPUQuota) updateThrottle() error {
	// update suppression times and duration
	// if data cannot be obtained from cpu.stat, the value remains unchanged.
	c.preThrottle = c.curThrottle
	cs, err := c.GetCgroupAttr(cpuStatKey).CPUStat()
	if err != nil {
		return err
	}
	c.curThrottle = cs
	return nil
}

func (c *CPUQuota) updateQuota() error {
	c.quotaDelta = 0
	curQuota, err := c.GetCgroupAttr(cpuQuotaKey).Int64()
	if err != nil {
		return err
	}
	c.curQuota = curQuota
	return nil
}

func (c *CPUQuota) updateUsage() error {
	latest, err := c.GetCgroupAttr(cpuAcctUsageKey).Int64()
	if err != nil {
		return err
	}
	c.cpuUsages = append(c.cpuUsages, cpuUsage{timestamp: time.Now().UnixNano(), usage: latest})
	// ensure that the CPU usage of the container does not exceed the upper limit.
	if len(c.cpuUsages) >= numberOfRestrictedCycles {
		c.cpuUsages = c.cpuUsages[1:]
	}
	return nil
}

func writeQuota(mountPoint string, paths []string, delta int64) error {
	type cgroupQuotaPair struct {
		h     *cgroup.Hierarchy
		value int64
	}
	var (
		writed []cgroupQuotaPair
		save   = func(mountPoint, path string, delta int64) error {
			h := cgroup.NewHierarchy(mountPoint, path)
			curQuota, err := h.GetCgroupAttr(cpuQuotaKey).Int64()
			if err != nil {
				return fmt.Errorf("error getting cgroup %v quota: %v", path, err)
			}
			if curQuota == -1 {
				return nil
			}

			nextQuota := curQuota + delta
			if err := h.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(nextQuota)); err != nil {
				return fmt.Errorf("error setting cgroup %v quota (%v to %v): %v", path, curQuota, nextQuota, err)
			}
			writed = append(writed, cgroupQuotaPair{h: h, value: curQuota})
			return nil
		}

		fallback = func() {
			for _, w := range writed {
				if err := w.h.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(w.value)); err != nil {
					fmt.Printf("error recovering cgroup %v quota %v\n", w.h.Path, w.value)
				}
			}
		}
	)

	if delta > 0 {
		// update the parent cgroup first, then update the child cgroup
		for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
			paths[i], paths[j] = paths[j], paths[i]
		}
	}

	for _, path := range paths {
		if err := save(mountPoint, path, delta); err != nil {
			fallback()
			return err
		}
	}
	return nil
}

// writeQuota use to modify quota for cgroup
func (c *CPUQuota) writeQuota() error {
	var (
		delta    = c.nextQuota - c.curQuota
		paths    []string
		fullPath = c.Path
	)
	if delta == 0 {
		return nil
	}
	// the upper cgroup needs to be updated synchronously
	if len(fullPath) == 0 {
		return fmt.Errorf("invalid cgroup path: %v", fullPath)
	}
	for {
		/*
			a non-slash start will end up with .
			start with a slash and end up with slash
		*/
		if fullPath == "." || fullPath == "/" || fullPath == "kubepods" || fullPath == "/kubepods" {
			break
		}
		paths = append(paths, fullPath)
		fullPath = path.Dir(fullPath)
	}
	if len(paths) == 0 {
		return fmt.Errorf("empty cgroup path")
	}
	if err := writeQuota(c.MountPoint, paths, delta); err != nil {
		return err
	}
	c.curQuota = c.nextQuota
	return nil
}

func (c *CPUQuota) recoverQuota() error {
	// period ranges from 1000(us) to 1000000(us) and does not overflow.
	c.nextQuota = int64(c.cpuLimit * float64(c.period))
	return c.writeQuota()
}

func appendErr(errs error, err error) error {
	if errs == nil {
		return err
	}
	if err == nil {
		return errs
	}
	errStr1 := errs.Error()
	errStr2 := err.Error()
	return fmt.Errorf("%s \n* %s", errStr1, errStr2)
}
