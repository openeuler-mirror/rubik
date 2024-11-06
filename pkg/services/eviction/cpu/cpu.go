// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
//	http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-11-04
// Description: This file is used for cpu evict service

// Package cpu provide cpu eviction services
package cpu

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/lib/cpu/quotaturbo"
	"isula.org/rubik/pkg/services/helper"
)

// usage is used to save CPU utilization data
type usage struct {
	timestamp time.Time
	cpuStats  quotaturbo.ProcStat
}

// Controller is used to collect CPU utilization
type Controller struct {
	sync.RWMutex
	conf   *Config
	usages []usage
	block  int32
}

// fromConfig generates CPU controller based on configuration
func fromConfig(name string, f helper.ConfigHandler) (*Controller, error) {
	var conf = newConfig()
	if err := f(name, conf); err != nil {
		return nil, err
	}
	// If the user does not set windows, or sets windows to 0, the default windows interval is 2 times
	if conf.Windows == defaultWindows {
		conf.Windows = 2 * conf.Interval
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return &Controller{
		conf: conf,
	}, nil
}

// Start loop collects data and performs eviction
func (c *Controller) Start(ctx context.Context, evictor func() error) {
	wait.Until(
		func() {
			c.collect()
			if atomic.LoadInt32(&c.block) == 1 {
				return
			}
			if !c.assertWithinLimit() {
				return
			}
			if err := evictor(); err != nil {
				log.Errorf("failed to execute cpuevict %v", err)
				return
			}
			// if the eviction is successful, it will enter the cool-down period.
			atomic.StoreInt32(&c.block, 1) // prevent future evictions
			// start a goroutine to reset the blocking flag
			go func() {
				time.Sleep(time.Duration(c.conf.Cooldown) * time.Second) // wait for cool down time
				atomic.StoreInt32(&c.block, 0)                           // allow future evictions
			}()
		},
		time.Second*time.Duration(c.conf.Interval),
		ctx.Done())
}

// collect records current CPU utilization
func (c *Controller) collect() {
	stat, err := quotaturbo.GetProcStat()
	if err != nil {
		log.Errorf("failed to get cpu usage: %v", err)
		return
	}
	c.purgeExpiredRecords()
	c.Lock()
	c.usages = append(c.usages, usage{timestamp: time.Now(), cpuStats: stat})
	c.Unlock()
}

// purgeExpiredRecords is used to clear data outside the window period
func (c *Controller) purgeExpiredRecords() {
	c.Lock()
	defer c.Unlock()
	var (
		now      = time.Now()
		latest   = len(c.usages)
		duration = time.Second * time.Duration(c.conf.Windows+1)
	)
	for i, usage := range c.usages {
		if now.Sub(usage.timestamp) < duration {
			latest = i
			break
		}
	}
	c.usages = c.usages[latest:]
}

// averageUsage calculates the average CPU usage within the specified time window.
func (c *Controller) averageUsage() float64 {
	const (
		minUsageLen = 2
		format      = "2006-01-02 15:04:05"
	)
	c.RLock()
	defer c.RUnlock()
	if len(c.usages) < minUsageLen {
		log.Infof("failed to get node cpu usage at %v", time.Now().Format(format))
		return 0
	}
	util := quotaturbo.CalculateUtils(c.usages[0].cpuStats, c.usages[len(c.usages)-1].cpuStats)
	log.Debugf("get node cpu usage %v at %v", util, time.Now().Format(format))
	return util
}

func (c *Controller) assertWithinLimit() bool {
	return c.averageUsage() >= float64(c.conf.Threshold)
}

// Config returns the configuration
func (c *Controller) Config() interface{} {
	return c.conf
}
