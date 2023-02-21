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

// Package quotaturbo is for Quota Turbo
package quotaturbo

import (
	"fmt"
	"path"
	"time"

	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
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
	*typedef.ContainerInfo
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
func NewCPUQuota(ci *typedef.ContainerInfo) (*CPUQuota, error) {
	defaultQuota := ci.LimitResources[typedef.ResourceCPU] * float64(defaultCFSPeriodUs)
	cq := &CPUQuota{
		ContainerInfo:      ci,
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

	if err := cq.updatePeriod(); err != nil {
		return cq, err
	}

	if err := cq.updateThrottle(); err != nil {
		return cq, err
	}
	// The throttle data before and after the initialization is the same.
	cq.preThrottle = cq.curThrottle

	if err := cq.updateQuota(); err != nil {
		return cq, err
	}

	if err := cq.updateUsage(); err != nil {
		return cq, err
	}
	return cq, nil
}

func (c *CPUQuota) updatePeriod() error {
	us, err := c.ContainerInfo.GetCgroupAttr(cpuPeriodKey).Int64()
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
	cs, err := c.ContainerInfo.GetCgroupAttr(cpuStatKey).CPUStat()
	if err != nil {
		return err
	}
	c.curThrottle = cs
	return nil
}

func (c *CPUQuota) updateQuota() error {
	c.quotaDelta = 0
	curQuota, err := c.ContainerInfo.GetCgroupAttr(cpuQuotaKey).Int64()
	if err != nil {
		return err
	}
	c.curQuota = curQuota
	return nil
}

func (c *CPUQuota) updateUsage() error {
	latest, err := c.ContainerInfo.GetCgroupAttr(cpuAcctUsageKey).Int64()
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

func (c *CPUQuota) writePodQuota(delta int64) error {
	pod := &typedef.PodInfo{
		CgroupPath: path.Dir(c.CgroupPath),
	}
	podQuota, err := pod.GetCgroupAttr(cpuQuotaKey).Int64()
	if err == nil && podQuota == -1 {
		return nil
	}
	podQuota += delta
	return pod.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(podQuota))
}

func (c *CPUQuota) writeContainerQuota() error {
	return c.ContainerInfo.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(c.curQuota))
}

// WriteQuota use to modify quota for container
func (c *CPUQuota) WriteQuota() error {
	delta := c.nextQuota - c.curQuota
	tmp := c.curQuota
	c.curQuota = c.nextQuota
	if delta < 0 {
		// update container data first
		if err := c.writeContainerQuota(); err != nil {
			c.curQuota = tmp
			return fmt.Errorf("fail to write container's quota for container %v: %v", c.ID, err)
		}
		// then update the pod data
		if err := c.writePodQuota(delta); err != nil {
			// recover
			c.curQuota = tmp
			if recoverErr := c.writeContainerQuota(); recoverErr != nil {
				log.Errorf("fail to recover contaienr's quota for container %v: %v", c.ID, recoverErr)
			}
			return fmt.Errorf("fail to write pod's quota for container %v: %v", c.ID, err)
		}
	} else if delta > 0 {
		// update pod data first
		if err := c.writePodQuota(delta); err != nil {
			c.curQuota = tmp
			return fmt.Errorf("fail to write pod's quota for container %v: %v", c.ID, err)
		}
		// then update the container data
		if err := c.writeContainerQuota(); err != nil {
			// recover
			c.curQuota = tmp
			if recoverErr := c.writePodQuota(-delta); recoverErr != nil {
				log.Errorf("fail to recover pod's quota for container %v: %v", c.ID, recoverErr)
			}
			return fmt.Errorf("fail to write container's quota for container %v: %v", c.ID, err)
		}
	}
	return nil
}
