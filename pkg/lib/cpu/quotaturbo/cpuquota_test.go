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
// Description: This file is used for testing cpu_quota

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

// TestNewCPUQuota tests NewCPUQuota
func TestNewCPUQuota(t *testing.T) {
	const (
		cpuPeriodFile = "cpu.cfs_period_us"
		cpuQuotaFile  = "cpu.cfs_quota_us"
		cpuUsageFile  = "cpuacct.usage"
		cpuStatFile   = "cpu.stat"
		validStat     = `nr_periods 1
		nr_throttled 1
		throttled_time 1
		`
		throttleTime int64 = 1
		quota              = "200000"
		quotaValue   int64 = 200000
		period             = "100000"
		periodValue  int64 = 100000
		usage              = "1234567"
		usageValue   int64 = 1234567
	)

	var (
		cgPath = "kubepods/testPod1/testCon1"
		h      = &cgroup.Hierarchy{
			MountPoint: constant.TmpTestDir,
			Path:       cgPath,
		}

		contPath       = filepath.Join(constant.TmpTestDir, "cpu", cgPath, "")
		contPeriodPath = filepath.Join(contPath, cpuPeriodFile)
		contQuotaPath  = filepath.Join(contPath, cpuQuotaFile)
		contUsagePath  = filepath.Join(constant.TmpTestDir, "cpuacct", cgPath, cpuUsageFile)
		contStatPath   = filepath.Join(contPath, cpuStatFile)
	)

	try.RemoveAll(constant.TmpTestDir)
	try.MkdirAll(contPath, constant.DefaultDirMode)
	try.MkdirAll(path.Dir(contUsagePath), constant.DefaultDirMode)
	defer try.RemoveAll(constant.TmpTestDir)

	const cpuLimit = 2.0
	// absent of period file
	try.RemoveAll(contPeriodPath)
	_, err := NewCPUQuota(h, cpuLimit)
	assert.Error(t, err, "should lacking of period file")
	try.WriteFile(contPeriodPath, period)

	// absent of throttle file
	try.RemoveAll(contStatPath)
	_, err = NewCPUQuota(h, cpuLimit)
	assert.Error(t, err, "should lacking of throttle file")
	try.WriteFile(contStatPath, validStat)

	// absent of quota file
	try.RemoveAll(contQuotaPath)
	_, err = NewCPUQuota(h, cpuLimit)
	assert.Error(t, err, "should lacking of quota file")
	try.WriteFile(contQuotaPath, quota)

	// absent of usage file
	try.RemoveAll(contUsagePath)
	_, err = NewCPUQuota(h, cpuLimit)
	assert.Error(t, err, "should lacking of usage file")
	try.WriteFile(contUsagePath, usage)

	cq, err := NewCPUQuota(h, cpuLimit)
	assert.NoError(t, err)
	assert.Equal(t, usageValue, cq.cpuUsages[0].usage)
	assert.Equal(t, quotaValue, cq.curQuota)
	assert.Equal(t, periodValue, cq.period)

	cu := make([]cpuUsage, numberOfRestrictedCycles)
	for i := 0; i < numberOfRestrictedCycles; i++ {
		cu[i] = cpuUsage{}
	}
	cq.cpuUsages = cu
	assert.NoError(t, cq.updateUsage())
}

// TestCPUQuota_WriteQuota tests WriteQuota of CPUQuota
func TestCPUQuota_WriteQuota(t *testing.T) {
	const (
		largerQuota             = "200000"
		largerQuotaVal    int64 = 200000
		smallerQuota            = "100000"
		smallerQuotaVal   int64 = 100000
		unlimitedQuota          = "-1"
		unlimitedQuotaVal int64 = -1
		periodUs                = "100000"
		cpuPeriodFile           = "cpu.cfs_period_us"
		cpuQuotaFile            = "cpu.cfs_quota_us"
	)

	var (
		cgPath         = "kubepods/testPod1/testCon1"
		contPath       = filepath.Join(constant.TmpTestDir, "cpu", cgPath, "")
		podPeriodPath  = filepath.Join(filepath.Dir(contPath), cpuPeriodFile)
		podQuotaPath   = filepath.Join(filepath.Dir(contPath), cpuQuotaFile)
		contPeriodPath = filepath.Join(contPath, cpuPeriodFile)
		contQuotaPath  = filepath.Join(contPath, cpuQuotaFile)
		assertValue    = func(t *testing.T, paths []string, value string) {
			for _, p := range paths {
				data, err := util.ReadFile(p)
				assert.NoError(t, err)
				assert.Equal(t, value, strings.TrimSpace(string(data)))
			}
		}
	)

	try.RemoveAll(constant.TmpTestDir)
	defer try.RemoveAll(constant.TmpTestDir)

	type fields struct {
		Hierarchy *cgroup.Hierarchy
		curQuota  int64
		nextQuota int64
	}
	tests := []struct {
		name    string
		pre     func()
		fields  fields
		post    func(t *testing.T, cq *CPUQuota)
		wantErr bool
	}{
		{
			name: "TC1-empty cgroup path",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{},
				nextQuota: largerQuotaVal,
				curQuota:  smallerQuotaVal,
			},
			wantErr: true,
		},
		{
			name: "TC2-fail to get paths",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					Path: "/",
				},
				nextQuota: largerQuotaVal,
				curQuota:  smallerQuotaVal,
			},
			wantErr: true,
		},
		{
			name: "TC3-None of the paths exist",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.TmpTestDir,
					Path:       "kubepods/testPod1/testCon1",
				},
				nextQuota: largerQuotaVal,
				curQuota:  smallerQuotaVal,
			},
			wantErr: true,
		},
		{
			name: "TC4-Only pod path existed",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.TmpTestDir,
					Path:       "kubepods/testPod1/testCon1",
				},
				// write the pod first and then write the container
				nextQuota: largerQuotaVal,
				curQuota:  smallerQuotaVal,
			},
			pre: func() {
				try.WriteFile(podQuotaPath, smallerQuota)
				try.WriteFile(podPeriodPath, periodUs)
				try.RemoveAll(contQuotaPath)
				try.RemoveAll(contPeriodPath)
			},
			post: func(t *testing.T, cq *CPUQuota) {
				// Unable to write to container, so restore pod as it is
				assertValue(t, []string{podQuotaPath}, smallerQuota)
				assert.Equal(t, smallerQuotaVal, cq.curQuota)
			},
			wantErr: true,
		},
		{
			name: "TC5-success delta > 0",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.TmpTestDir,
					Path:       "kubepods/testPod1/testCon1",
				},
				nextQuota: largerQuotaVal,
				curQuota:  smallerQuotaVal,
			},
			pre: func() {
				try.WriteFile(podQuotaPath, smallerQuota)
				try.WriteFile(podPeriodPath, periodUs)
				try.WriteFile(contQuotaPath, smallerQuota)
				try.WriteFile(contPeriodPath, periodUs)
			},
			post: func(t *testing.T, cq *CPUQuota) {
				assertValue(t, []string{podQuotaPath, contQuotaPath}, largerQuota)
				assert.Equal(t, largerQuotaVal, cq.curQuota)
			},
			wantErr: false,
		},
		{
			name: "TC6-success delta < 0",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.TmpTestDir,
					Path:       "kubepods/testPod1/testCon1",
				},
				nextQuota: smallerQuotaVal,
				curQuota:  largerQuotaVal,
			},
			pre: func() {
				try.WriteFile(podQuotaPath, largerQuota)
				try.WriteFile(podPeriodPath, periodUs)
				try.WriteFile(contQuotaPath, largerQuota)
				try.WriteFile(contPeriodPath, periodUs)
			},
			post: func(t *testing.T, cq *CPUQuota) {
				assertValue(t, []string{podQuotaPath, contQuotaPath}, smallerQuota)
				assert.Equal(t, smallerQuotaVal, cq.curQuota)
			},
			wantErr: false,
		},
		{
			name: "TC6.1-success delta < 0 unlimited pod",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: constant.TmpTestDir,
					Path:       "kubepods/testPod1/testCon1",
				},
				nextQuota: smallerQuotaVal,
				curQuota:  largerQuotaVal,
			},
			pre: func() {
				try.WriteFile(podQuotaPath, unlimitedQuota)
				try.WriteFile(podPeriodPath, periodUs)
				try.WriteFile(contQuotaPath, largerQuota)
				try.WriteFile(contPeriodPath, periodUs)
			},
			post: func(t *testing.T, cq *CPUQuota) {
				assertValue(t, []string{contQuotaPath}, smallerQuota)
				assertValue(t, []string{podQuotaPath}, unlimitedQuota)
				assert.Equal(t, smallerQuotaVal, cq.curQuota)
			},
			wantErr: false,
		},
		{
			name: "TC7-success delta = 0",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{},
				nextQuota: smallerQuotaVal,
				curQuota:  smallerQuotaVal,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CPUQuota{
				Hierarchy: tt.fields.Hierarchy,
				curQuota:  tt.fields.curQuota,
				nextQuota: tt.fields.nextQuota,
			}
			if tt.pre != nil {
				tt.pre()
			}
			if err := c.writeQuota(); (err != nil) != tt.wantErr {
				t.Errorf("CPUQuota.WriteQuota() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.post != nil {
				tt.post(t, c)
			}
		})
	}
}
