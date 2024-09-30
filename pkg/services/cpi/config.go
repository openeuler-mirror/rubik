// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Kelu Ye
// Date: 2024-09-19
// Description: This file contains default configurations used in the CPU Interference Detection Service

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"time"

	"isula.org/rubik/pkg/core/typedef/cgroup"
)

var (
	// cpuQuotaKey is a cgroup key that refers to the CPU quota file.
	cpuQuotaKey = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_quota_us"}
	// cpuUsageKey is a cgroup key that refers to the CPU usage accounting file.
	cpuUsageKey = &cgroup.Key{SubSys: "cpu", FileName: "cpuacct.usage"}
)

const (
	// defaultCollectDur is the default interval for collecting CPI and CPU usage data.
	defaultCollectDur = time.Minute * 1
	// defaultSampleDur is the default sampling duration for collecting CPI and CPU usage data.
	defaultSampleDur = time.Second * 10
	// defaultExpireDur is the default interval for deleting expired data.
	defaultExpireDur = time.Minute * 30
	// defaultIdentifyDur is the default interval for detecting CPU interference.
	defaultIdentifyDur = time.Minute * 5
	// defaultWindowDur is the time window used to determine the correlation between suspected offline tasks and online tasks.
	defaultWindowDur = time.Minute * 10
	// defaultAntagonistMetric is the metric used to determine the interference of offline tasks on online tasks.
	defaultAntagonistMetric = 0.15
	// defaultLimitDur is the duration for which the antagonist is restricted from using the CPU.
	defaultLimitDur = time.Minute * 5
	// defaultLimitQuota is the CPU quota value assigned to the antagonist after interference is detected.
	defaultLimitQuota = "10000"
	//defaultMinCount is the minimum sample count required for a CPI Statistic to be considered valid
	defaultMinCount = 60
)
