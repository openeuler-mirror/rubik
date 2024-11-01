//go:build linux
// +build linux

// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2024-10-31
// Description: This file defines cadvisor config

package cadvisor

import (
	"time"

	"github.com/google/cadvisor/cache/memory"
	"github.com/google/cadvisor/container"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/utils/sysfs"
)

// Config is a set of parameters that control the startup of cadvisor
type Config struct {
	MemCache              *memory.InMemoryCache
	SysFs                 sysfs.SysFs
	IncludeMetrics        container.MetricSet
	MaxHousekeepingConfig manager.HouskeepingConfig
}

func (c *Config) Config() *Config {
	return c
}

type ConfigOpt func(args *Config)

func WithCacheAge(cacheAge time.Duration) ConfigOpt {
	return func(args *Config) {
		args.MemCache = memory.New(cacheAge, nil)
	}
}

func WithFs(fs sysfs.SysFs) ConfigOpt {
	return func(args *Config) {
		args.SysFs = fs
	}
}

func WithMetrics(metrics string) ConfigOpt {
	return func(args *Config) {
		var ms container.MetricSet
		ms.Set(metrics)
		args.IncludeMetrics = ms
	}
}

func WithHouseKeepingInterval(interval time.Duration) ConfigOpt {
	return func(args *Config) {
		args.MaxHousekeepingConfig.Interval = &interval
	}
}

func WithHouseKeepingDynamic(allowDynamic bool) ConfigOpt {
	return func(args *Config) {
		args.MaxHousekeepingConfig.AllowDynamic = &allowDynamic
	}
}

func NewConfig(opts ...ConfigOpt) *Config {
	var (
		allowDynamic = true
		interval     = time.Second
	)
	var conf = &Config{
		MaxHousekeepingConfig: manager.HouskeepingConfig{
			AllowDynamic: &allowDynamic,
			Interval:     &interval,
		},
		SysFs: sysfs.NewRealSysFs(),
	}
	for _, opt := range opts {
		opt(conf)
	}
	return conf
}
