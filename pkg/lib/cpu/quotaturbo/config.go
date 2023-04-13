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
// Date: 2023-03-07
// Description: This file is used for quota turbo config

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"

	"isula.org/rubik/pkg/common/constant"
)

const (
	defaultHighWaterMark     = 60
	defaultAlarmWaterMark    = 80
	defaultElevateLimit      = 1.0
	defaultSlowFallbackRatio = 0.1
	defaultCPUFloatingLimit  = 10.0
)

// Config defines configuration of QuotaTurbo
type Config struct {
	/*
		If the CPU utilization exceeds HighWaterMark, it will trigger a slow fall,
	*/
	HighWaterMark int `json:"highWaterMark,omitempty"`
	/*
		If the CPU utilization exceeds the Alarm WaterMark, it will trigger a fast fallback;
		otherwise it will trigger a slow increase
	*/
	AlarmWaterMark int `json:"alarmWaterMark,omitempty"`
	/*
		Cgroup Root indicates the mount point of the system Cgroup, the default is /sys/fs/cgroup
	*/
	CgroupRoot string `json:"cgroupRoot,omitempty"`
	/*
		ElevateLimit is the maximum percentage(%) of the total amount of
		a single promotion to the total amount of nodes
		Default is 1.0
	*/
	ElevateLimit float64 `json:"elevateLimit,omitempty"`
	/*
		Slow Fallback Ratio is used to control the rate of slow fallback. Default is 0.1
	*/
	SlowFallbackRatio float64 `json:"slowFallbackRatio,omitempty"`
	/*
		CPUFloatingLimit indicates the Upper Percentage Change of the CPU utilization of the node
		within the specified time period.
		Only when the floating rate is lower than the upper limit can the quota be increased,
		and the decrease is not limited
		Default is 10.0
	*/
	CPUFloatingLimit float64 `json:"cpuFloatingLimit,omitempty"`
}

// Option is an option provided by the Client for setting parameters
type Option func(c *Config) error

// NewConfig returns a quota Turbo config instance with default values
func NewConfig() *Config {
	return &Config{
		HighWaterMark:     defaultHighWaterMark,
		AlarmWaterMark:    defaultAlarmWaterMark,
		CgroupRoot:        constant.DefaultCgroupRoot,
		ElevateLimit:      defaultElevateLimit,
		SlowFallbackRatio: defaultSlowFallbackRatio,
		CPUFloatingLimit:  defaultCPUFloatingLimit,
	}
}

// validateWaterMark verifies that the WaterMark is set correctly
func (c *Config) validateWaterMark() error {
	const minQuotaTurboWaterMark, maxQuotaTurboWaterMark = 0, 100
	outOfRange := func(num int) bool {
		return num < minQuotaTurboWaterMark || num > maxQuotaTurboWaterMark
	}
	if c.AlarmWaterMark <= c.HighWaterMark || outOfRange(c.HighWaterMark) || outOfRange(c.AlarmWaterMark) {
		return fmt.Errorf("alarmWaterMark >= highWaterMark, both of which ranges from 0 to 100")
	}
	return nil
}
