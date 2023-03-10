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
// Date: 2023-02-16
// Description: This file is used for testing statusstore.go

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

// TestIsAdjustmentAllowed tests isAdjustmentAllowed
func TestIsAdjustmentAllowed(t *testing.T) {
	const contPath1 = "kubepods/testPod1/testCon1"

	try.RemoveAll(constant.TmpTestDir)
	defer try.RemoveAll(constant.TmpTestDir)

	tests := []struct {
		h        *cgroup.Hierarchy
		cpuLimit float64
		pre      func()
		post     func()
		name     string
		want     bool
	}{
		{
			name: "TC1-allow adjustment",
			h: &cgroup.Hierarchy{
				MountPoint: constant.TmpTestDir,
				Path:       contPath1,
			},
			cpuLimit: float64(runtime.NumCPU()) - 1,
			pre: func() {
				try.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1, "cpu.cfs_quota_us"),
					constant.DefaultFileMode)
			},
			post: func() {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1))
			},
			want: true,
		},
		{
			name: "TC2-cgroup path is not existed",
			h: &cgroup.Hierarchy{
				MountPoint: constant.TmpTestDir,
				Path:       contPath1,
			},
			cpuLimit: float64(runtime.NumCPU()) - 1,
			pre: func() {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1))
			},
			want: false,
		},
		{
			name: "TC3-cpulimit = 0",
			h: &cgroup.Hierarchy{
				MountPoint: constant.TmpTestDir,
				Path:       contPath1,
			},
			cpuLimit: 0,
			pre: func() {
				try.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1, "cpu.cfs_quota_us"),
					constant.DefaultFileMode)
			},
			post: func() {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1))
			},
			want: false,
		},
		{
			name: "TC4-cpulimit over max",
			h: &cgroup.Hierarchy{
				MountPoint: constant.TmpTestDir,
				Path:       contPath1,
			},
			cpuLimit: float64(runtime.NumCPU()) + 1,
			pre: func() {
				try.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1, "cpu.cfs_quota_us"),
					constant.DefaultFileMode)
			},
			post: func() {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1))
			},
			want: false,
		},
		{
			name: "TC5-cpurequest over max",
			h: &cgroup.Hierarchy{
				MountPoint: constant.TmpTestDir,
				Path:       contPath1,
			},
			cpuLimit: 0,
			pre: func() {
				try.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1, "cpu.cfs_quota_us"),
					constant.DefaultFileMode)
			},
			post: func() {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath1))
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre()
			}
			assert.Equal(t, isAdjustmentAllowed(tt.h, tt.cpuLimit), tt.want)
			if tt.post != nil {
				tt.post()
			}
		})
	}
}

// TestStatusStore_RemoveCgroup tests RemoveCgroup of StatusStore
func TestStatusStore_RemoveCgroup(t *testing.T) {
	const (
		podPath  = "kubepods/testPod1"
		contPath = "kubepods/testPod1/testCon1"
	)
	type fields struct {
		Config    *Config
		cpuQuotas map[string]*CPUQuota
	}
	type args struct {
		cgroupPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		pre     func()
		post    func(t *testing.T, d *StatusStore)
	}{
		{
			name: "TC1-empty cgroupPath",
			args: args{
				cgroupPath: "",
			},
			fields: fields{
				cpuQuotas: make(map[string]*CPUQuota),
			},
			wantErr: false,
		},
		{
			name: "TC2-cgroupPath is not existed",
			args: args{
				cgroupPath: contPath,
			},
			fields: fields{
				cpuQuotas: map[string]*CPUQuota{
					contPath: {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.TmpTestDir,
							Path:       contPath,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TC3-cgroupPath existed but can not set",
			args: args{
				cgroupPath: contPath,
			},
			fields: fields{
				Config: &Config{
					CgroupRoot: constant.TmpTestDir,
				},
				cpuQuotas: map[string]*CPUQuota{
					contPath: {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.TmpTestDir,
							Path:       contPath,
						},
						curQuota:  100000,
						nextQuota: 200000,
					},
				},
			},
			pre: func() {
				try.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", contPath), constant.DefaultDirMode)
			},
			post: func(t *testing.T, d *StatusStore) {
				try.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", contPath))
			},
			wantErr: true,
		},
		{
			name: "TC4-remove cgroupPath successfully",
			args: args{
				cgroupPath: contPath,
			},
			fields: fields{
				cpuQuotas: map[string]*CPUQuota{
					contPath: {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.TmpTestDir,
							Path:       contPath,
						},
						cpuLimit:  2,
						period:    100000,
						curQuota:  250000,
						nextQuota: 240000,
					},
				},
			},
			pre: func() {
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, "cpu.cfs_quota_us"), "250000")
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", podPath, "cpu.cfs_quota_us"), "-1")
			},
			post: func(t *testing.T, d *StatusStore) {
				val := strings.TrimSpace(try.ReadFile(
					filepath.Join(constant.TmpTestDir, "cpu", contPath, "cpu.cfs_quota_us")).String())
				assert.Equal(t, "200000", val)
				val = strings.TrimSpace(try.ReadFile(
					filepath.Join(constant.TmpTestDir, "cpu", podPath, "cpu.cfs_quota_us")).String())
				assert.Equal(t, "-1", val)
				assert.Equal(t, 0, len(d.cpuQuotas))
				try.RemoveAll(constant.TmpTestDir)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &StatusStore{
				Config:    tt.fields.Config,
				cpuQuotas: tt.fields.cpuQuotas,
			}
			if tt.pre != nil {
				tt.pre()
			}
			if err := d.RemoveCgroup(tt.args.cgroupPath); (err != nil) != tt.wantErr {
				t.Errorf("StatusStore.RemoveCgroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.post != nil {
				tt.post(t, d)
			}
		})
	}
}

// TestStatusStore_AddCgroup tests AddCgroup of StatusStore
func TestStatusStore_AddCgroup(t *testing.T) {
	const (
		contPath      = "kubepods/testPod1/testCon1"
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
	type fields struct {
		Config    *Config
		cpuQuotas map[string]*CPUQuota
	}
	type args struct {
		cgroupPath string
		cpuLimit   float64
		cpuRequest float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		pre     func(t *testing.T, d *StatusStore)
		post    func(t *testing.T, d *StatusStore)
	}{
		{
			name: "TC1-empty cgroup path",
			args: args{
				cgroupPath: "",
			},
			fields: fields{
				cpuQuotas: make(map[string]*CPUQuota),
			},
			wantErr: true,
		},
		{
			name: "TC2-empty cgroup mount point",
			args: args{
				cgroupPath: contPath,
			},
			fields: fields{
				Config: &Config{
					CgroupRoot: "",
				},
			},
			wantErr: true,
		},
		{
			name: "TC3-cgroup not allow to adjust",
			args: args{
				cgroupPath: contPath,
				cpuLimit:   3,
			},
			fields: fields{
				Config: &Config{
					CgroupRoot: constant.TmpTestDir,
				},
			},
			wantErr: true,
		},
		{
			name: "TC4-failed to create CPUQuota",
			args: args{
				cgroupPath: contPath,
				cpuLimit:   3,
			},
			fields: fields{
				Config: &Config{
					CgroupRoot: constant.TmpTestDir,
				},
				cpuQuotas: make(map[string]*CPUQuota),
			},
			pre: func(t *testing.T, d *StatusStore) {
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuPeriodFile), period)
			},
			post: func(t *testing.T, d *StatusStore) {
				try.RemoveAll(constant.TmpTestDir)
			},
			wantErr: true,
		},
		{
			name: "TC5-add successfully",
			args: args{
				cgroupPath: contPath,
				cpuLimit:   2,
			},
			fields: fields{
				Config: &Config{
					CgroupRoot: constant.TmpTestDir,
				},
				cpuQuotas: make(map[string]*CPUQuota),
			},
			pre: func(t *testing.T, d *StatusStore) {
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuPeriodFile), period)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuQuotaFile), quota)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpuacct", contPath, cpuUsageFile), usage)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuStatFile), stat)
			},
			post: func(t *testing.T, d *StatusStore) {
				assert.Equal(t, 1, len(d.cpuQuotas))
				try.RemoveAll(constant.TmpTestDir)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &StatusStore{
				Config:    tt.fields.Config,
				cpuQuotas: tt.fields.cpuQuotas,
			}
			if tt.pre != nil {
				tt.pre(t, d)
			}
			if err := d.AddCgroup(tt.args.cgroupPath, tt.args.cpuLimit); (err != nil) != tt.wantErr {
				t.Errorf("StatusStore.AddCgroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.post != nil {
				tt.post(t, d)
			}
		})
	}
}

// TestStatusStoreGetLastCPUUtil tests getLastCPUUtil of StatusStore
func TestStatusStore_getLastCPUUtil(t *testing.T) {
	// 1. empty CPU Utils
	d := &StatusStore{}
	t.Run("TC1-empty CPU Util", func(t *testing.T) {
		util := float64(0.0)
		assert.Equal(t, util, d.getLastCPUUtil())
	})
	// 2. CPU Utils
	cpuUtil20 := 20
	d = &StatusStore{cpuUtils: []cpuUtil{{
		util: float64(cpuUtil20),
	}}}
	t.Run("TC2-CPU Util is 20", func(t *testing.T) {
		util := float64(20.0)
		assert.Equal(t, util, d.getLastCPUUtil())
	})
}

// TestQuotaTurboUpdateCPUUtils tests updateCPUUtils of QuotaTurbo and NewProcStat
func TestStatusStore_updateCPUUtils(t *testing.T) {
	status := NewStatusStore()
	// 1. obtain the cpu usage for the first time
	if err := status.updateCPUUtils(); err != nil {
		assert.NoError(t, err)
	}
	num1 := 1
	assert.Equal(t, num1, len(status.cpuUtils))
	// 2. obtain the cpu usage for the second time
	if err := status.updateCPUUtils(); err != nil {
		assert.NoError(t, err)
	}
	num2 := 2
	assert.Equal(t, num2, len(status.cpuUtils))
	// 3. obtain the cpu usage after 1 minute
	var minuteTimeDelta int64 = 60000000001
	status.cpuUtils[0].timestamp -= minuteTimeDelta
	if err := status.updateCPUUtils(); err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, num2, len(status.cpuUtils))
}

// TestStatusStore_updateCPUQuotas tests updateCPUQuotas of StatusStore
func TestStatusStore_updateCPUQuotas(t *testing.T) {
	const (
		contPath      = "kubepods/testPod1/testCon1"
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
	type fields struct {
		Config    *Config
		cpuQuotas map[string]*CPUQuota
		cpuUtils  []cpuUtil
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		pre     func()
		post    func()
	}{
		{
			name: "TC1-fail to get CPUQuota",
			fields: fields{
				Config: &Config{
					CgroupRoot: constant.TmpTestDir,
				},
				cpuQuotas: map[string]*CPUQuota{
					contPath: {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.TmpTestDir,
							Path:       contPath,
						},
					},
				},
				cpuUtils: make([]cpuUtil, 0),
			},
			wantErr: true,
		},
		{
			name: "TC2-update successfully",
			fields: fields{
				Config: &Config{
					CgroupRoot: constant.TmpTestDir,
				},
				cpuQuotas: map[string]*CPUQuota{
					contPath: {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: constant.TmpTestDir,
							Path:       contPath,
						},
					},
				},
				cpuUtils: make([]cpuUtil, 0),
			},
			pre: func() {
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuPeriodFile), period)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuQuotaFile), quota)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpuacct", contPath, cpuUsageFile), usage)
				try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuStatFile), stat)
			},
			post: func() {
				try.RemoveAll(constant.TmpTestDir)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &StatusStore{
				Config:    tt.fields.Config,
				cpuQuotas: tt.fields.cpuQuotas,
				cpuUtils:  tt.fields.cpuUtils,
			}
			if tt.pre != nil {
				tt.pre()
			}
			if err := d.updateCPUQuotas(); (err != nil) != tt.wantErr {
				t.Errorf("StatusStore.update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.post != nil {
				tt.post()
			}
		})
	}
}

// TestStatusStore_writeQuota tests writeQuota of StatusStore
func TestStatusStore_writeQuota(t *testing.T) {
	const contPath = "kubepods/testPod1/testCon1"
	tests := []struct {
		name      string
		cpuQuotas map[string]*CPUQuota
		wantErr   bool
	}{
		{
			name: "TC1-empty cgroup path",
			cpuQuotas: map[string]*CPUQuota{
				contPath: {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: constant.TmpTestDir,
						Path:       contPath,
					},
					curQuota:  100000,
					nextQuota: 200000,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &StatusStore{
				cpuQuotas: tt.cpuQuotas,
			}
			if err := d.writeQuota(); (err != nil) != tt.wantErr {
				t.Errorf("StatusStore.writeQuota() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
