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
// Date: 2023-03-09
// Description: This file is used for testing quota turbo client

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

// TestClient_AdjustQuota tests AdjustQuota of Client
func TestClient_AdjustQuota(t *testing.T) {
	const (
		contPath      = "kubepods/testPod1/testCon1"
		podPath       = "kubepods/testPod1"
		cpuPeriodFile = "cpu.cfs_period_us"
		cpuQuotaFile  = "cpu.cfs_quota_us"
		cpuUsageFile  = "cpuacct.usage"
		cpuStatFile   = "cpu.stat"
		stat          = `nr_periods 1
		nr_throttled 1
		throttled_time 1
		`
		quota  = "200000"
		period = "100000"
		usage  = "1234567"
	)
	tests := []struct {
		name    string
		wantErr bool
		pre     func(t *testing.T, c *Client)
		post    func(t *testing.T)
	}{
		{
			name:    "TC1-empty CPUQuotas",
			wantErr: false,
		},
		{
			name: "TC2-fail to updateCPUQuota causing absent of path",
			pre: func(t *testing.T, c *Client) {
				c.SetCgroupRoot(constant.TmpTestDir)
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath))
				try.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", contPath), constant.DefaultDirMode)

				assert.Equal(t, 0, len(c.cpuQuotas))
				c.AddCgroup(contPath, float64(runtime.NumCPU()))
				c.cpuQuotas[contPath] = &CPUQuota{
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: c.CgroupRoot,
						Path:       contPath,
					},
					cpuLimit:    float64(runtime.NumCPU()) - 1,
					curThrottle: &cgroup.CPUStat{},
					preThrottle: &cgroup.CPUStat{},
				}
				assert.Equal(t, 1, len(c.cpuQuotas))
			},
			post: func(t *testing.T) {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath))
			},
			wantErr: true,
		},
		{
			name: "TC3-success",
			pre: func(t *testing.T, c *Client) {
				c.SetCgroupRoot(constant.TmpTestDir)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuPeriodFile), period)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuQuotaFile), quota)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpuacct", contPath, cpuUsageFile), usage)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuStatFile), stat)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", podPath, cpuPeriodFile), period)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", podPath, cpuQuotaFile), quota)
				assert.NoError(t, c.AddCgroup(contPath, float64(runtime.NumCPU())-1))
				assert.Equal(t, 1, len(c.cpuQuotas))
			},
			post: func(t *testing.T) {
				try.RemoveAll(constant.TmpTestDir)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient()
			if tt.pre != nil {
				tt.pre(t, c)
			}
			if err := c.AdjustQuota(); (err != nil) != tt.wantErr {
				t.Errorf("Client.AdjustQuota() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
