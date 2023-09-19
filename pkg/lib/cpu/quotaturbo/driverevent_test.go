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
// Description: This file is used for testing driverevent.go

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"math"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// TestEventDriverElevate tests elevate of EventDriver
func TestEventDriverElevate(t *testing.T) {
	var elevateTests = []struct {
		status     *StatusStore
		judgements func(t *testing.T, status *StatusStore)
		name       string
	}{
		{
			name: "TC1 - CPU usage >= the alarmWaterMark.",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 60,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon1": {},
				},
				cpuUtils: []cpuUtil{
					{
						util: 90,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				var delta float64 = 0
				conID := "testCon1"
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC2 - the container is not suppressed.",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 70,
					ElevateLimit:   defaultElevateLimit,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon2": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod2/testCon2",
						},
						// currently not suppressed
						curThrottle: &cgroup.CPUStat{
							NrThrottled:   1,
							ThrottledTime: 10,
						},
						preThrottle: &cgroup.CPUStat{
							NrThrottled:   1,
							ThrottledTime: 10,
						},
						period: 100000,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 60,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				var delta float64 = 0
				conID := "testCon2"
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC3 - increase the quota of the suppressed container",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 60,
					ElevateLimit:   defaultElevateLimit,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon3": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod3/testCon3",
						},
						curThrottle: &cgroup.CPUStat{
							NrThrottled:   50,
							ThrottledTime: 200000,
						},
						preThrottle: &cgroup.CPUStat{
							NrThrottled:   40,
							ThrottledTime: 100000,
						},
						period: 100000,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 40,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				conID := "testCon3"
				c := status.cpuQuotas[conID]
				coefficient := math.Min(float64(0.0001),
					util.PercentageToDecimal(status.ElevateLimit)*float64(runtime.NumCPU())) /
					float64(0.0001)
				delta := coefficient * float64(0.0001) * float64(c.period)
				assert.True(t, status.cpuQuotas[conID].quotaDelta == delta)
			},
		},
	}

	e := &EventDriver{}
	for _, tt := range elevateTests {
		t.Run(tt.name, func(t *testing.T) {
			e.elevate(tt.status)
			tt.judgements(t, tt.status)
		})
	}
}

// TestSlowFallback tests slowFallback of EventDriver
func TestSlowFallback(t *testing.T) {
	var slowFallBackTests = []struct {
		status     *StatusStore
		judgements func(t *testing.T, status *StatusStore)
		name       string
	}{
		{
			name: "TC1-CPU usage <= the highWaterMark.",
			status: &StatusStore{
				Config: &Config{
					HighWaterMark:     60,
					SlowFallbackRatio: defaultSlowFallbackRatio,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon4": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod4/testCon4",
						},
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 40,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				conID := "testCon4"
				var delta float64 = 0
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC2-the container is suppressed.",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark:    80,
					HighWaterMark:     50,
					SlowFallbackRatio: defaultSlowFallbackRatio,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon5": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod5/testCon5",
						},
						cpuLimit: 1,
						curThrottle: &cgroup.CPUStat{
							NrThrottled: 10,
						},
						preThrottle: &cgroup.CPUStat{
							NrThrottled: 0,
						},
						period:   100000,
						curQuota: 200000,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 70,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				var delta float64 = 0
				conID := "testCon5"
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC3-decrease the quota of the uncompressed containers",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark:    90,
					HighWaterMark:     40,
					SlowFallbackRatio: defaultSlowFallbackRatio,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon6": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod6/testCon6",
						},
						cpuLimit: 2,
						// currently not suppressed
						curThrottle: &cgroup.CPUStat{
							NrThrottled:   10,
							ThrottledTime: 100000,
						},
						preThrottle: &cgroup.CPUStat{
							NrThrottled:   10,
							ThrottledTime: 100000,
						},
						period:   100000,
						curQuota: 400000,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 60.0,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				conID := "testCon6"
				c := status.cpuQuotas[conID]
				coefficient := (status.getLastCPUUtil() - float64(status.HighWaterMark)) /
					float64(status.AlarmWaterMark-status.HighWaterMark) * status.SlowFallbackRatio
				delta := coefficient *
					((float64(c.cpuLimit) * float64(c.period)) - float64(c.curQuota))
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
	}
	e := &EventDriver{}
	for _, tt := range slowFallBackTests {
		t.Run(tt.name, func(t *testing.T) {
			e.slowFallback(tt.status)
			tt.judgements(t, tt.status)
		})
	}
}

// TestFastFallback tests fastFallback of EventDriver
func TestFastFallback(t *testing.T) {
	var fastFallBackTests = []struct {
		status     *StatusStore
		judgements func(t *testing.T, status *StatusStore)
		name       string
	}{
		{
			name: "TC1-CPU usage <= the AlarmWaterMark.",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 30,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon7": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod7/testCon7",
						},
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 10,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				conID := "testCon7"
				var delta float64 = 0
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC2-the quota of container is not increased.",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 30,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon8": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod8/testCon8",
						},
						cpuLimit: 1,
						period:   100,
						curQuota: 100,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 48,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				var delta float64 = 0
				conID := "testCon8"
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC3-decrease the quota of the containers",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 65,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon9": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod9/testCon9",
						},
						cpuLimit: 3,
						period:   10000,
						curQuota: 40000,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 90,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				conID := "testCon9"
				c := status.cpuQuotas[conID]
				delta := util.PercentageToDecimal(float64(status.AlarmWaterMark)-status.getLastCPUUtil()) *
					float64(runtime.NumCPU()) * float64(c.period)
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
	}
	e := &EventDriver{}
	for _, tt := range fastFallBackTests {
		t.Run(tt.name, func(t *testing.T) {
			e.fastFallback(tt.status)
			tt.judgements(t, tt.status)
		})
	}
}

// TestSharpFluctuates tests sharpFluctuates
func TestSharpFluctuates(t *testing.T) {
	const cpuUtil90 = 90
	var sharpFluctuatesTests = []struct {
		status *StatusStore
		want   bool
		name   string
	}{
		{
			name: "TC1-the cpu changes rapidly",
			status: &StatusStore{
				Config: &Config{
					CPUFloatingLimit: defaultCPUFloatingLimit,
				},
				cpuUtils: []cpuUtil{
					{
						util: cpuUtil90,
					},
					{
						util: cpuUtil90 - defaultCPUFloatingLimit - 1,
					},
				},
			},
			want: true,
		},
		{
			name: "TC2-the cpu changes steadily",
			status: &StatusStore{
				Config: &Config{
					CPUFloatingLimit: defaultCPUFloatingLimit,
				},
				cpuUtils: []cpuUtil{
					{
						util: cpuUtil90,
					},
					{
						util: cpuUtil90 - defaultCPUFloatingLimit + 1,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range sharpFluctuatesTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, sharpFluctuates(tt.status) == tt.want)
		})
	}
}

// TestEventDriverAdjustQuota tests adjustQuota of EventDriver
func TestEventDriverAdjustQuota(t *testing.T) {
	var eDriverAdjustQuotaTests = []struct {
		status     *StatusStore
		judgements func(t *testing.T, status *StatusStore)
		name       string
	}{
		{
			name: "TC1-no promotion",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 80,
					HighWaterMark:  73,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon10": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod10/testCon10",
						},
						cpuLimit: 1,
						period:   80,
						curQuota: 100,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 1,
					},
					{
						util: -defaultCPUFloatingLimit,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				var delta float64 = 0
				conID := "testCon10"
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
		{
			name: "TC2-make a promotion",
			status: &StatusStore{
				Config: &Config{
					AlarmWaterMark: 97,
					HighWaterMark:  73,
				},
				cpuQuotas: map[string]*CPUQuota{
					"testCon11": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.DefaultCgroupRoot,
							Path:       "kubepods/testPod11/testCon11",
						},
						cpuLimit: 2,
						curThrottle: &cgroup.CPUStat{
							NrThrottled:   1,
							ThrottledTime: 200,
						},
						preThrottle: &cgroup.CPUStat{
							NrThrottled:   0,
							ThrottledTime: 100,
						},
						period:   2000,
						curQuota: 5000,
					},
				},
				cpuUtils: []cpuUtil{
					{
						util: 10,
					},
				},
			},
			judgements: func(t *testing.T, status *StatusStore) {
				conID := "testCon11"
				c := status.cpuQuotas[conID]
				coefficient := math.Min(float64(0.00005), util.PercentageToDecimal(status.ElevateLimit)*
					float64(runtime.NumCPU())) / float64(0.00005)
				delta := coefficient * float64(0.00005) * float64(c.period)
				assert.Equal(t, delta, status.cpuQuotas[conID].quotaDelta)
			},
		},
	}
	e := &EventDriver{}
	for _, tt := range eDriverAdjustQuotaTests {
		t.Run(tt.name, func(t *testing.T) {
			e.adjustQuota(tt.status)
			tt.judgements(t, tt.status)
		})
	}
}

// TestGetMaxQuota tests getMaxQuota
func TestGetMaxQuota(t *testing.T) {
	var getMaxQuotaTests = []struct {
		cq         *CPUQuota
		judgements func(t *testing.T, cq *CPUQuota)
		name       string
	}{
		{
			name: "TC1-empty cpu usage",
			cq: &CPUQuota{
				heightLimit: 100,
				cpuUsages:   []cpuUsage{},
			},
			judgements: func(t *testing.T, cq *CPUQuota) {
				var res float64 = 100
				assert.Equal(t, res, getMaxQuota(cq))
			},
		},
		{
			name: "TC2-The remaining value is less than 3 times the upper limit.",
			cq: &CPUQuota{
				cpuUsages: []cpuUsage{
					{100000, 100000},
					{200000, 200000},
				},
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.DefaultCgroupRoot,
					Path:       "kubepods/testPod1/testCon1",
				},
				cpuLimit:    4,
				period:      100,
				heightLimit: 800,
			},
			judgements: func(t *testing.T, cq *CPUQuota) {
				const res = 400 + float64(400*700)/float64(3*800)
				assert.Equal(t, res, getMaxQuota(cq))
			},
		},
		{
			name: "TC3-The remaining value is greater than 3 times the limit height.",
			cq: &CPUQuota{
				cpuUsages: []cpuUsage{
					{10000, 0},
					{20000, 0},
					{30000, 0},
					{40000, 0},
					{50000, 0},
					{60000, 0},
					{70000, 0},
					{80000, 100},
				},
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.DefaultCgroupRoot,
					Path:       "kubepods/testPod1/testCon1",
				},
				cpuLimit:    1,
				period:      100,
				heightLimit: 200,
			},
			judgements: func(t *testing.T, cq *CPUQuota) {
				var res float64 = 200
				assert.Equal(t, res, getMaxQuota(cq))
			},
		},
		{
			name: "TC4-The remaining value is less than the initial value.",
			cq: &CPUQuota{
				cpuUsages: []cpuUsage{
					{100, 0},
					{200, 1000000},
				},
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.DefaultCgroupRoot,
					Path:       "kubepods/testPod1/testCon1",
				},
				cpuLimit:    10,
				period:      10,
				heightLimit: 150,
			},
			judgements: func(t *testing.T, cq *CPUQuota) {
				var res float64 = 100
				assert.Equal(t, res, getMaxQuota(cq))
			},
		},
	}
	for _, tt := range getMaxQuotaTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.judgements(t, tt.cq)
		})
	}
}
