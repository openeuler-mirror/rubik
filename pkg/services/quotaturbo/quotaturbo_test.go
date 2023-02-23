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
// Description: This file is used for test quota turbo

// Package quotaturbo is for Quota Turbo
package quotaturbo

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/common/constant"
	Log "isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/podmanager"
)

// TestQuotaTurbo_Validatec test Validate function
func TestQuotaTurbo_Validate(t *testing.T) {
	tests := []struct {
		name     string
		NodeData *NodeData
		wantErr  bool
	}{
		{
			name: "TC1-alarmWaterMark is less or equal to highWaterMark",
			NodeData: &NodeData{Config: &Config{
				HighWaterMark:  100,
				AlarmWaterMark: 60,
				SyncInterval:   1000,
			}},
			wantErr: true,
		},
		{
			name: "TC2-highWater mark exceed the max quota turbo water mark(100)",
			NodeData: &NodeData{Config: &Config{
				HighWaterMark:  1000,
				AlarmWaterMark: 100000,
				SyncInterval:   1000,
			}},
			wantErr: true,
		},
		{
			name: "TC3-sync interval out of range(100-10000)",
			NodeData: &NodeData{Config: &Config{
				HighWaterMark:  60,
				AlarmWaterMark: 80,
				SyncInterval:   1,
			}},
			wantErr: true,
		},
		{
			name: "TC4-normal case",
			NodeData: &NodeData{Config: &Config{
				HighWaterMark:  60,
				AlarmWaterMark: 100,
				SyncInterval:   1000,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := &QuotaTurbo{
				NodeData: tt.NodeData,
			}
			if err := qt.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("QuotaTurbo.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
				contList = []*typedef.ContainerInfo{
					fooCont,
					barCont,
				}
				pod = &typedef.PodInfo{
					UID:        "testPod1",
					CgroupPath: "kubepods/testPod1",
					IDContainersMap: map[string]*typedef.ContainerInfo{
						fooCont.ID: fooCont,
						barCont.ID: barCont,
					},
				}
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

			mkCgDirs(contList)
			defer rmCgDirs(contList)

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

// TestQuotaTurbo_AdjustQuota tests AdjustQuota function
func TestQuotaTurbo_AdjustQuota(t *testing.T) {
	type fields struct {
		NodeData *NodeData
		Driver   Driver
		pm       *podmanager.PodManager
	}
	type args struct {
		cc map[string]*typedef.ContainerInfo
	}
	var (
		fooCont = &typedef.ContainerInfo{
			Name:       "Foo",
			ID:         "testCon1",
			CgroupPath: "kubepods/testPod1/testCon1",
			LimitResources: typedef.ResourceMap{
				typedef.ResourceCPU: 2,
			},
			RequestResources: typedef.ResourceMap{
				typedef.ResourceCPU: 2,
			},
		}
		pod1 = &typedef.PodInfo{
			UID:        "testPod1",
			CgroupPath: "kubepods/testPod1",
			IDContainersMap: map[string]*typedef.ContainerInfo{
				fooCont.ID: fooCont,
			},
		}
		pod2 = &typedef.PodInfo{
			UID:             "testPod2",
			CgroupPath:      "kubepods/testPod2",
			IDContainersMap: map[string]*typedef.ContainerInfo{},
		}
	)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TC1-empty data",
			args: args{
				cc: map[string]*typedef.ContainerInfo{},
			},
			fields: fields{
				NodeData: &NodeData{
					Config: &Config{
						AlarmWaterMark: 80,
						HighWaterMark:  60,
					},
					containers: make(map[string]*CPUQuota),
				},
				Driver: &EventDriver{},
				pm: &podmanager.PodManager{
					Pods: &podmanager.PodCache{
						Pods: map[string]*typedef.PodInfo{
							pod1.UID: pod1,
							pod2.UID: pod2,
						},
					},
				},
			},
		},
		{
			name: "TC2-existed data",
			args: args{
				cc: map[string]*typedef.ContainerInfo{
					"testCon1": fooCont,
				},
			},
			fields: fields{
				NodeData: &NodeData{
					Config: &Config{
						AlarmWaterMark: 80,
						HighWaterMark:  60,
					},
					containers: make(map[string]*CPUQuota),
				},
				Driver: &EventDriver{},
				pm: &podmanager.PodManager{
					Pods: &podmanager.PodCache{
						Pods: map[string]*typedef.PodInfo{
							pod1.UID: pod1,
							pod2.UID: pod2,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := &QuotaTurbo{
				NodeData: tt.fields.NodeData,
				Driver:   tt.fields.Driver,
				Viewer:   tt.fields.pm,
			}
			conts := []*typedef.ContainerInfo{}
			for _, p := range tt.fields.pm.Pods.Pods {
				for _, c := range p.IDContainersMap {
					conts = append(conts, c)
				}
			}
			mkCgDirs(conts)
			defer rmCgDirs(conts)

			qt.AdjustQuota(tt.args.cc)
		})
	}
}

// TestNewQuotaTurbo tests NewQuotaTurbo
func TestNewQuotaTurbo(t *testing.T) {
	testName := "TC1-test otherv functions"
	t.Run(testName, func(t *testing.T) {
		got := NewQuotaTurbo()
		got.SetupLog(&Log.EmptyLog{})
		assert.Equal(t, moduleName, got.ID())
		got.Viewer = &podmanager.PodManager{
			Pods: &podmanager.PodCache{
				Pods: map[string]*typedef.PodInfo{
					"testPod1": {
						UID:        "testPod1",
						CgroupPath: "kubepods/testPod1",
						Annotations: map[string]string{
							constant.QuotaAnnotationKey: "true",
						},
					},
				},
			},
		}

		ctx, cancle := context.WithCancel(context.Background())
		go got.Run(ctx)
		time.Sleep(time.Second)
		cancle()
	})
}

// TestQuotaTurbo_PreStart tests PreStart
func TestQuotaTurbo_PreStart(t *testing.T) {
	var (
		pm = &podmanager.PodManager{
			Pods: &podmanager.PodCache{
				Pods: map[string]*typedef.PodInfo{
					"testPod1": {
						UID:             "testPod1",
						CgroupPath:      "kubepods/testPod1",
						IDContainersMap: make(map[string]*typedef.ContainerInfo),
					},
				},
			},
		}
		qt = &QuotaTurbo{}
	)
	testName := "TC1- test Prestart"
	t.Run(testName, func(t *testing.T) {
		qt.PreStart(pm)
	})
}
