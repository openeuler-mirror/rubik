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
// Date: 2023-03-11
// Description: This file is used for testing quotaturbo.go

package quotaturbo

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/podmanager"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/test/try"
)

func sameQuota(t *testing.T, path string, want int64) bool {
	const cfsQuotaUsFileName = "cpu.cfs_quota_us"
	data, err := util.ReadSmallFile(filepath.Join(path, cfsQuotaUsFileName))
	if err != nil {
		assert.NoError(t, err)
		return false
	}
	quota, err := util.ParseInt64(strings.ReplaceAll(string(data), "\n", ""))
	if err != nil {
		assert.NoError(t, err)
		return false
	}
	if quota != want {
		return false
	}
	return true
}

// TestQuotaTurbo_Terminate tests Terminate function
func TestQuotaTurbo_Terminate(t *testing.T) {
	const (
		fooContName   = "Foo"
		barContName   = "Bar"
		podUID        = "testPod1"
		wrongPodQuota = "600000"
		wrongFooQuota = "300000"
		wrongBarQuota = "200000"
		podPath       = "/sys/fs/cgroup/cpu/kubepods/testPod1/"
		fooPath       = "/sys/fs/cgroup/cpu/kubepods/testPod1/testCon1"
		barPath       = "/sys/fs/cgroup/cpu/kubepods/testPod1/testCon2"
	)

	var (
		fooCont = &typedef.ContainerInfo{
			Name:           fooContName,
			ID:             "testCon1",
			CgroupPath:     "kubepods/testPod1/testCon1",
			LimitResources: make(typedef.ResourceMap),
		}
		barCont = &typedef.ContainerInfo{
			Name:           barContName,
			ID:             "testCon2",
			CgroupPath:     "kubepods/testPod1/testCon2",
			LimitResources: make(typedef.ResourceMap),
		}
		pod = &typedef.PodInfo{
			UID:        "testPod1",
			CgroupPath: "kubepods/testPod1",
			IDContainersMap: map[string]*typedef.ContainerInfo{
				fooCont.ID: fooCont,
				barCont.ID: barCont,
			},
		}
		tests = []struct {
			postfunc    func(t *testing.T)
			fooCPULimit float64
			barCPULimit float64
			name        string
		}{
			{
				name:        "TC1-one unlimited container is existed",
				fooCPULimit: 2,
				barCPULimit: 0,
				postfunc: func(t *testing.T) {
					var (
						unlimited       int64 = -1
						correctFooQuota int64 = 200000
					)
					assert.True(t, sameQuota(t, podPath, unlimited))
					assert.True(t, sameQuota(t, fooPath, correctFooQuota))
					assert.True(t, sameQuota(t, barPath, unlimited))
				},
			},
			{
				name:        "TC2-all containers are unlimited",
				fooCPULimit: 2,
				barCPULimit: 1,
				postfunc: func(t *testing.T) {
					var (
						correctPodQuota int64 = 300000
						correctFooQuota int64 = 200000
						correctBarQuota int64 = 100000
					)
					assert.True(t, sameQuota(t, podPath, correctPodQuota))
					assert.True(t, sameQuota(t, fooPath, correctFooQuota))
					assert.True(t, sameQuota(t, barPath, correctBarQuota))
				},
			},
			{
				name:        "TC3-all containers are limited",
				fooCPULimit: 0,
				barCPULimit: 0,
				postfunc: func(t *testing.T) {
					var unLimited int64 = -1
					assert.True(t, sameQuota(t, podPath, unLimited))
					assert.True(t, sameQuota(t, fooPath, unLimited))
					assert.True(t, sameQuota(t, barPath, unLimited))
				},
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				pm = &podmanager.PodManager{
					Pods: &podmanager.PodCache{
						Pods: map[string]*typedef.PodInfo{
							podUID: pod,
						},
					},
				}
				qt = &QuotaTurbo{
					Viewer: pm,
				}
			)

			try.MkdirAll(fooPath, constant.DefaultDirMode)
			try.MkdirAll(barPath, constant.DefaultDirMode)
			defer func() {
				try.RemoveAll(podPath)
			}()

			fooCont.LimitResources[typedef.ResourceCPU] = tt.fooCPULimit
			barCont.LimitResources[typedef.ResourceCPU] = tt.barCPULimit

			assert.NoError(t, pod.SetCgroupAttr(cpuQuotaKey, wrongPodQuota))
			assert.NoError(t, fooCont.SetCgroupAttr(cpuQuotaKey, wrongFooQuota))
			assert.NoError(t, barCont.SetCgroupAttr(cpuQuotaKey, wrongBarQuota))
			qt.Terminate(pm)
			tt.postfunc(t)
		})
	}
}

// TestQuotaTurbo_PreStart tests PreStart
func TestQuotaTurbo_PreStart(t *testing.T) {
	const (
		fooContName = "Foo"
		podUID      = "testPod1"
		podPath     = "/sys/fs/cgroup/cpu/kubepods/testPod1/"
	)
	var (
		fooCont = &typedef.ContainerInfo{
			Name:           fooContName,
			ID:             "testCon1",
			CgroupPath:     "kubepods/testPod1/testCon1",
			LimitResources: make(typedef.ResourceMap),
		}
		pm = &podmanager.PodManager{
			Pods: &podmanager.PodCache{
				Pods: map[string]*typedef.PodInfo{
					podUID: {
						UID:        podUID,
						CgroupPath: "kubepods/testPod1",
						IDContainersMap: map[string]*typedef.ContainerInfo{
							fooCont.ID: fooCont,
						},
					},
				},
			},
		}
		name = "quotaturbo"
		qt   = NewQuotaTurbo(name)
	)
	testName := "TC1- test Prestart"
	t.Run(testName, func(t *testing.T) {
		qt.PreStart(pm)
		try.MkdirAll(podPath, constant.DefaultDirMode)
		defer try.RemoveAll(podPath)
		qt.PreStart(pm)
		pm.Pods.Pods[podUID].IDContainersMap[fooCont.ID].
			LimitResources[typedef.ResourceCPU] = math.Min(1, float64(runtime.NumCPU())-1)
		qt.PreStart(pm)
	})
}

// TestConfig_Validate test Validate function
func TestConfig_Validate(t *testing.T) {
	type fields struct {
		HighWaterMark  int
		AlarmWaterMark int
		SyncInterval   int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TC1-alarmWaterMark is less or equal to highWaterMark",
			fields: fields{
				HighWaterMark:  100,
				AlarmWaterMark: 60,
				SyncInterval:   1000,
			},
			wantErr: true,
		},
		{
			name: "TC2-highWater mark exceed the max quota turbo water mark(100)",
			fields: fields{
				HighWaterMark:  1000,
				AlarmWaterMark: 100000,
				SyncInterval:   1000,
			},
			wantErr: true,
		},
		{
			name: "TC3-sync interval out of range(100-10000)",
			fields: fields{
				HighWaterMark:  60,
				AlarmWaterMark: 80,
				SyncInterval:   1,
			},
			wantErr: true,
		},
		{
			name: "TC4-normal case",
			fields: fields{
				HighWaterMark:  60,
				AlarmWaterMark: 100,
				SyncInterval:   1000,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &Config{
				HighWaterMark:  tt.fields.HighWaterMark,
				AlarmWaterMark: tt.fields.AlarmWaterMark,
				SyncInterval:   tt.fields.SyncInterval,
			}
			if err := conf.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestQuotaTurbo_SetConfig tests SetConfig
func TestQuotaTurbo_SetConfig(t *testing.T) {
	const name = "quotaturbo"
	type args struct {
		f helper.ConfigHandler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TC1-error function",
			args: args{
				f: func(configName string, d interface{}) error { return fmt.Errorf("error occures") },
			},
			wantErr: true,
		},
		{
			name: "TC2-success",
			args: args{
				f: func(configName string, d interface{}) error { return nil },
			},
			wantErr: false,
		},
		{
			name: "TC3-invalid config",
			args: args{
				f: func(configName string, d interface{}) error {
					c, ok := d.(*Config)
					if !ok {
						t.Error("fial to convert config")
					}
					c.AlarmWaterMark = 101
					return nil
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := NewQuotaTurbo(name)
			if err := qt.SetConfig(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("QuotaTurbo.SetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestQuotaTurbo_Other tests other function
func TestQuotaTurbo_Other(t *testing.T) {
	const name = "quotaturbo"
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "TC1-test other",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &QuotaTurboFactory{ObjName: name}
			f.Name()
			instance, err := f.NewObj()
			assert.NoError(t, err)
			qt, ok := instance.(*QuotaTurbo)
			if !ok {
				t.Error("fial to convert QuotaTurbo")
			}
			if got := qt.IsRunner(); got != tt.want {
				t.Errorf("QuotaTurbo.IsRunner() = %v, want %v", got, tt.want)
			}
			qt.ID()

		})
	}
}

// TestQuotaTurbo_AdjustQuota tests AdjustQuota
func TestQuotaTurbo_AdjustQuota(t *testing.T) {
	const (
		name          = "quotaturbo"
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
		minCPU = 2
	)
	var (
		fooCont = &typedef.ContainerInfo{
			Name:       "Foo",
			ID:         "testCon1",
			CgroupPath: "kubepods/testPod1/testCon1",
			LimitResources: typedef.ResourceMap{
				typedef.ResourceCPU: math.Min(minCPU, float64(runtime.NumCPU()-1)),
			},
		}
		barCont = &typedef.ContainerInfo{
			Name:       "Bar",
			ID:         "testCon2",
			CgroupPath: "kubepods/testPod2/testCon2",
			LimitResources: typedef.ResourceMap{
				typedef.ResourceCPU: math.Min(minCPU, float64(runtime.NumCPU()-1)),
			},
		}
		preEnv = func(contPath string) {
			try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuPeriodFile), period)
			try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuQuotaFile), quota)
			try.WriteFile(filepath.Join(constant.TmpTestDir, "cpuacct", contPath, cpuUsageFile), usage)
			try.WriteFile(filepath.Join(constant.TmpTestDir, "cpu", contPath, cpuStatFile), stat)
		}
	)
	type args struct {
		conts map[string]*typedef.ContainerInfo
	}
	tests := []struct {
		name string
		args args
		pre  func(t *testing.T, qt *QuotaTurbo)
		post func(t *testing.T)
	}{
		{
			name: "TC1-fail add foo container & remove bar container successfully",
			args: args{
				conts: map[string]*typedef.ContainerInfo{
					fooCont.ID: fooCont,
				},
			},
			pre: func(t *testing.T, qt *QuotaTurbo) {
				preEnv(barCont.CgroupPath)
				assert.NoError(t, qt.client.AddCgroup(barCont.CgroupPath, barCont.LimitResources[typedef.ResourceCPU]))
				assert.Equal(t, 1, len(qt.client.GetAllCgroup()))
			},
			post: func(t *testing.T) {
				try.RemoveAll(constant.TmpTestDir)
			},
		},
		{
			name: "TC2-no container add  & remove bar container successfully",
			args: args{
				conts: map[string]*typedef.ContainerInfo{
					fooCont.ID: fooCont,
					barCont.ID: barCont,
				},
			},
			pre: func(t *testing.T, qt *QuotaTurbo) {
				preEnv(barCont.CgroupPath)
				preEnv(fooCont.CgroupPath)
				assert.NoError(t, qt.client.AddCgroup(barCont.CgroupPath, barCont.LimitResources[typedef.ResourceCPU]))
				assert.NoError(t, qt.client.AddCgroup(fooCont.CgroupPath, fooCont.LimitResources[typedef.ResourceCPU]))
				const cgroupLen = 2
				assert.Equal(t, cgroupLen, len(qt.client.GetAllCgroup()))
			},
			post: func(t *testing.T) {
				try.RemoveAll(constant.TmpTestDir)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := NewQuotaTurbo(name)
			qt.client.SetCgroupRoot(constant.TmpTestDir)
			if tt.pre != nil {
				tt.pre(t, qt)
			}
			qt.AdjustQuota(tt.args.conts)
			if tt.post != nil {
				tt.post(t)
			}
		})
	}
}

// TestQuotaTurbo_Run tests run
func TestQuotaTurbo_Run(t *testing.T) {
	const name = "quotaturbo"
	var fooPod = &typedef.PodInfo{
		Name:       "Foo",
		UID:        "testPod1",
		CgroupPath: "kubepods/testPod1",
		Annotations: map[string]string{
			constant.QuotaAnnotationKey: "true",
		},
	}
	tests := []struct {
		name string
	}{
		{
			name: "TC1-run and cancel",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := NewQuotaTurbo(name)
			pm := &podmanager.PodManager{
				Pods: &podmanager.PodCache{
					Pods: map[string]*typedef.PodInfo{
						fooPod.UID: fooPod,
					},
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			qt.Viewer = pm
			go qt.Run(ctx)
			const sleepTime = time.Millisecond * 200
			time.Sleep(sleepTime)
			cancel()
		})
	}
}
