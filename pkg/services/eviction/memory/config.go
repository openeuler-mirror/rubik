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

// Package memory provide memory eviction services
package memory

import (
	"fmt"
	"math"
)

const (
	// parameter value range
	minInterval  uint16 = 1
	maxInterval  uint16 = 3600
	minThreshold uint8  = 1
	maxThreshold uint8  = 99
	minCooldown  int    = 1
	maxCooldown  int    = math.MaxInt64 - 1

	// default value
	defaultInterval  uint16 = minInterval
	defaultCooldown  int    = 4
	defaultThreshold uint8  = 0
)

// Config is memory Evcit service configuration
type Config struct {
	Interval  uint16 `json:"interval,omitempty"`
	Threshold uint8  `json:"threshold,omitempty"`
	Cooldown  int    `json:"cooldown,omitempty"`
}

// newConfig returns default memory Evcit configuration
func newConfig() *Config {
	return &Config{
		Interval:  defaultInterval,
		Threshold: defaultThreshold,
		Cooldown:  defaultCooldown,
	}
}

// validate verifies that the memory Evcit parameter is set correctly
func (conf *Config) validate() error {
	if conf.Interval < minInterval || conf.Interval > maxInterval {
		return fmt.Errorf("interval should in the range [%v, %v]", minInterval, maxInterval)
	}
	if conf.Threshold == defaultThreshold {
		return fmt.Errorf("threshold must be set")
	}
	if conf.Threshold < minThreshold || conf.Threshold > maxThreshold {
		return fmt.Errorf("threshold should in the range [%v, %v]", minThreshold, maxThreshold)
	}
	if conf.Cooldown < minCooldown || conf.Cooldown > maxCooldown {
		return fmt.Errorf("cooldown should in the range [%v, %v]", minCooldown, maxCooldown)
	}
	return nil
}
