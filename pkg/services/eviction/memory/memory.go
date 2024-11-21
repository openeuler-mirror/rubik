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
// Date: 2024-11-06
// Description: This file is used for memory evict service

// Package memory provide memory eviction service
package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/services/helper"
)

// Controller is used to collect Memory utilization
type Controller struct {
	sync.RWMutex
	conf  *Config
	block int32
}

// fromConfig generates Memory Controller based on configuration
func fromConfig(name string, f helper.ConfigHandler) (*Controller, error) {
	var conf = newConfig()
	if err := f(name, conf); err != nil {
		return nil, err
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return &Controller{
		conf: conf,
	}, nil
}

// Start loop collects data and performs eviction
func (c *Controller) Start(ctx context.Context, evictor func(func() bool) error) {
	wait.Until(
		func() {
			if atomic.LoadInt32(&c.block) == 1 {
				return
			}
			if err := evictor(c.assertWithinLimit); err != nil {
				log.Errorf("failed to execute memory evict %v", err)
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

func (c *Controller) assertWithinLimit() bool {
	const (
		format = "2006-01-02 15:04:05"
	)
	// Get memory information
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("failed to get memory util at %v: %v", time.Now().Format(format), err)
		return false
	}
	if v.UsedPercent >= float64(c.conf.Threshold) {
		log.Infof("Memory exceeded: %v%%", v.UsedPercent)
		return true
	}
	return false
}

// Config returns the configuration
func (c *Controller) Config() interface{} {
	return c.conf
}
