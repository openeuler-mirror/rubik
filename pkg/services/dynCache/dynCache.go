// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2023-02-21
// Description: This file is cache limit service

// Package dyncache is the service used for cache limit setting
package dyncache

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/services/helper"
)

// default value
const (
	defaultAdInt        = 1000
	defaultPerfDur      = 1000
	defaultLowL3        = 20
	defaultMidL3        = 30
	defaultHighL3       = 50
	defaultLowMB        = 10
	defaultMidMB        = 30
	defaultHighMB       = 50
	defaultMaxMiss      = 20
	defaultMinMiss      = 10
	defaultResctrlDir   = "/sys/fs/resctrl"
	defaultNumaNodeDir  = "/sys/devices/system/node"
	defaultPidNameSpace = "/proc/self/ns/pid"
	modeStatic          = "static"
	modeDynamic         = "dynamic"
)

// boundary value
const (
	minPercent = 10
	maxPercent = 100
	// minimum perf duration, unit ms
	minPerfDur = 10
	// maximum perf duration, unit ms
	maxPerfDur = 10000
	// min adjust interval, unit ms
	minAdjustInterval = 10
	// max adjust interval, unit ms
	maxAdjustInterval = 10000
)

// MultiLvlPercent define multi level percentage
type MultiLvlPercent struct {
	Low  int `json:"low,omitempty"`
	Mid  int `json:"mid,omitempty"`
	High int `json:"high,omitempty"`
}

// Config is cache limit config
type Config struct {
	// DefaultLimitMode is default cache limit method(static)
	DefaultLimitMode string `json:"defaultLimitMode,omitempty"`
	// DefaultResctrlDir is the path of resctrl control group
	// the resctrl dir is not supposed to be altered or exposed
	DefaultResctrlDir string `json:"-"`
	// DefaultPidNameSpace is the pid namespace used for test whether share host pid
	// the value is not supposed to be altered or exposed
	DefaultPidNameSpace string `json:"-"`
	// AdjustInterval is dynamic cache adjust interval time
	AdjustInterval int `json:"adjustInterval,omitempty"`
	// PerfDuration is online pod perf dection duration time
	PerfDuration int `json:"perfDuration,omitempty"`
	// L3Percent is L3 cache percent for each level
	L3Percent MultiLvlPercent `json:"l3Percent,omitempty"`
	// MemBandPercent is memory bandwidth percent for each level
	MemBandPercent MultiLvlPercent `json:"memBandPercent,omitempty"`
}

// DynCache is cache limit service structure
type DynCache struct {
	helper.ServiceBase
	*Config
	Attr   *Attr
	Viewer api.Viewer
	Name   string `json:"-"`
}

// Attr is cache limit attribute differ from config
type Attr struct {
	// NumaNodeDir is the path for numa node
	NumaNodeDir string
	// NumaNum stores numa number on physical machine
	NumaNum int
	// L3PercentDynamic stores l3 percent for dynamic cache limit setting
	// the value could be changed
	L3PercentDynamic int
	// MemBandPercentDynamic stores memory band percent for dynamic cache limit setting
	// the value could be changed
	MemBandPercentDynamic int
	// MaxMiss is the maximum value of cache miss
	MaxMiss int
	// MinMiss is the minimum value of cache miss
	MinMiss int
}

// DynCacheFactory is the factory os dyncache.
type DynCacheFactory struct {
	ObjName string
}

// Name to get the dyncache factory name.
func (i DynCacheFactory) Name() string {
	return "DynCacheFactory"
}

// NewObj to create object of dyncache.
func (i DynCacheFactory) NewObj() (interface{}, error) {
	return NewDynCache(i.ObjName), nil
}

// NewDynCache return cache limit instance with default settings
func NewDynCache(name string) *DynCache {
	return &DynCache{
		Name: name,
		Attr: &Attr{
			NumaNodeDir: defaultNumaNodeDir,
			MaxMiss:     defaultMaxMiss,
			MinMiss:     defaultMinMiss,
		},
		Config: &Config{
			DefaultLimitMode:    modeStatic,
			DefaultResctrlDir:   defaultResctrlDir,
			DefaultPidNameSpace: defaultPidNameSpace,
			AdjustInterval:      defaultAdInt,
			PerfDuration:        defaultPerfDur,
			L3Percent: MultiLvlPercent{
				Low:  defaultLowL3,
				Mid:  defaultMidL3,
				High: defaultHighL3,
			},
			MemBandPercent: MultiLvlPercent{
				Low:  defaultLowMB,
				Mid:  defaultMidMB,
				High: defaultHighMB,
			},
		},
	}
}

// PreStart will do some pre-setting actions
func (c *DynCache) PreStart(viewer api.Viewer) error {
	c.Viewer = viewer

	if err := c.InitCacheLimitDir(); err != nil {
		return err
	}
	return nil
}

// ID returns service's name
func (c *DynCache) ID() string {
	return c.Name
}

// Run implement service run function
func (c *DynCache) Run(ctx context.Context) {
	go wait.Until(c.SyncCacheLimit, time.Second, ctx.Done())
	wait.Until(c.StartDynamic, time.Millisecond*time.Duration(c.Config.AdjustInterval), ctx.Done())
}

// Validate validate service's config
func (c *DynCache) Validate() error {
	defaultLimitMode := c.Config.DefaultLimitMode
	if defaultLimitMode != modeStatic && defaultLimitMode != modeDynamic {
		return fmt.Errorf("invalid cache limit mode: %s, should be %s or %s",
			c.Config.DefaultLimitMode, modeStatic, modeDynamic)
	}
	if c.Config.AdjustInterval < minAdjustInterval || c.Config.AdjustInterval > maxAdjustInterval {
		return fmt.Errorf("adjust interval %d out of range [%d,%d]",
			c.Config.AdjustInterval, minAdjustInterval, maxAdjustInterval)
	}
	if c.Config.PerfDuration < minPerfDur || c.Config.PerfDuration > maxPerfDur {
		return fmt.Errorf("perf duration %d out of range [%d,%d]", c.Config.PerfDuration, minPerfDur, maxPerfDur)
	}
	for _, per := range []int{
		c.Config.L3Percent.Low, c.Config.L3Percent.Mid,
		c.Config.L3Percent.High, c.Config.MemBandPercent.Low,
		c.Config.MemBandPercent.Mid, c.Config.MemBandPercent.High} {
		if per < minPercent || per > maxPercent {
			return fmt.Errorf("cache limit percentage %d out of range [%d,%d]", per, minPercent, maxPercent)
		}
	}
	if c.Config.L3Percent.Low > c.Config.L3Percent.Mid || c.Config.L3Percent.Mid > c.Config.L3Percent.High {
		return fmt.Errorf("cache limit config L3Percent does not satisfy constraint low<=mid<=high")
	}
	if c.Config.MemBandPercent.Low > c.Config.MemBandPercent.Mid || c.Config.MemBandPercent.Mid > c.Config.MemBandPercent.High {
		return fmt.Errorf("cache limit config MemBandPercent does not satisfy constraint low<=mid<=high")
	}
	return nil
}
