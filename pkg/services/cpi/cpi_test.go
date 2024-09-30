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
// Description: This file is used for testing cpi.go

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/podmanager"
)

var (
	fooOnlineCon = &typedef.ContainerInfo{
		Name: "onlineCon",
		ID:   "onlineCon",
		Hierarchy: cgroup.Hierarchy{
			MountPoint: "/",
			Path:       "online/testCon1"},
		LimitResources: make(typedef.ResourceMap),
	}
	fooOfflineCon = &typedef.ContainerInfo{
		Name: "offlineCon",
		ID:   "onlineCon",
		Hierarchy: cgroup.Hierarchy{
			MountPoint: "/",
			Path:       "offline/testCon1"},
		LimitResources: make(typedef.ResourceMap),
	}
	fooOnlinePod = &typedef.PodInfo{
		Name: "onlinePod",
		UID:  "onlinePod",
		Hierarchy: cgroup.Hierarchy{
			MountPoint: "/",
			Path:       "online",
		},
		Annotations: map[string]string{
			constant.CpiAnnotationKey: "online",
		},
		IDContainersMap: map[string]*typedef.ContainerInfo{
			fooOnlineCon.ID: fooOnlineCon,
		},
	}
	fooOfflinePod = &typedef.PodInfo{
		Name: "offlinePod",
		UID:  "offlinePod",
		Hierarchy: cgroup.Hierarchy{
			MountPoint: "/",
			Path:       "offline",
		},
		Annotations: map[string]string{
			constant.CpiAnnotationKey: "offline",
		},
		IDContainersMap: map[string]*typedef.ContainerInfo{
			fooOfflineCon.ID: fooOfflineCon,
		},
	}
	podWithNoAnno = &typedef.PodInfo{
		Name: "pod",
		UID:  "pod",
		Hierarchy: cgroup.Hierarchy{
			MountPoint: "/",
			Path:       "pod",
		},
		IDContainersMap: map[string]*typedef.ContainerInfo{
			fooOfflineCon.ID: fooOfflineCon,
		},
	}
	podWithErrorAnno = &typedef.PodInfo{
		Name: "offlinePod",
		UID:  "offlinePod",
		Hierarchy: cgroup.Hierarchy{
			MountPoint: "/",
			Path:       "offline",
		},
		Annotations: map[string]string{
			constant.CpiAnnotationKey: "error",
		},
		IDContainersMap: map[string]*typedef.ContainerInfo{
			fooOfflineCon.ID: fooOfflineCon,
		},
	}
)

// TestCpiServiceRun tests Run function
func TestCpiServiceRun(t *testing.T) {
	const name = "cpi"
	tests := []struct {
		name string
	}{
		{
			name: "Test CPI service run and cancel",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt := newCpiService(name)
			pm := &podmanager.PodManager{
				Pods: &podmanager.PodCache{
					Pods: map[string]*typedef.PodInfo{
						fooOnlinePod.UID: fooOnlinePod,
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

// TestCpiServicePreStart tests PreStart function
func TestCpiServicePreStart(t *testing.T) {
	_, err := perf.CgroupStat(cgroup.AbsoluteCgroupPath("perf_event", "", ""), time.Millisecond, cpiConf)
	if err != nil {
		return
	}
	createFile := func(dir string, fileName string, content string) {
		if err := os.MkdirAll(dir, constant.DefaultDirMode); err != nil {
			t.Errorf("error creating temp dir: %v", err)
		}
		file, err := os.Create(path.Join(dir, fileName))
		file.Chmod(constant.DefaultFileMode)
		if err != nil {
			t.Errorf("error creating quota file: %v", err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Errorf("error writing to quota file: %v", err)
		}
		file.Close()
	}
	var (
		pm = &podmanager.PodManager{
			Pods: &podmanager.PodCache{
				Pods: map[string]*typedef.PodInfo{
					fooOnlinePod.UID:  fooOnlinePod,
					fooOfflinePod.UID: fooOfflinePod,
				},
			},
		}
		name     = "Cpi"
		service  = newCpiService(name)
		fileName = "cpu.cfs_quota_us"
	)

	createFile(path.Join("/cpu", fooOfflinePod.MountPoint, fooOfflinePod.Path), fileName, "10000")
	createFile(path.Join("/cpu", fooOfflineCon.MountPoint, fooOfflineCon.Path), fileName, "10000")
	defer os.RemoveAll("/cpu")
	testName := "cpi- test Prestart"
	t.Run(testName, func(t *testing.T) {
		service.PreStart(pm)
		if len(service.onlineTasks) != 1 || len(service.offlineTasks) != 1 {
			t.Errorf("PreStart failed: expected 1 online task and 1 offline task, got %d online tasks and %d offline tasks", len(service.onlineTasks), len(service.offlineTasks))
		}
	})
}

// TestCpiServiceAddPod tests AddPod function
func TestCpiServiceAddPod(t *testing.T) {
	tests := []struct {
		name    string
		podInfo *typedef.PodInfo
		wantErr bool
	}{
		{
			name:    "Add onlinePod",
			podInfo: fooOnlinePod,
			wantErr: false,
		},
		{
			name:    "Add offlinePod",
			podInfo: fooOfflinePod,
			wantErr: false,
		},
		{
			name:    "cpi test add errorAnoPod",
			podInfo: podWithErrorAnno,
			wantErr: false,
		},
		{
			name:    "cpi test add no annotation pod",
			podInfo: podWithNoAnno,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newCpiService("Cpi")
			if err := service.AddPod(tt.podInfo); (err != nil) != tt.wantErr {
				t.Errorf("CpiService.AddPod() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

// TestCpiServiceDeletePod tests DeletePod function
func TestCpiServiceDeletePod(t *testing.T) {
	tests := []struct {
		name    string
		podInfo *typedef.PodInfo
		wantErr bool
	}{
		{
			name:    "cpi test delete onlinePod",
			podInfo: fooOnlinePod,
			wantErr: false,
		},
		{
			name:    "cpi test delete offlinePod",
			podInfo: fooOfflinePod,
			wantErr: false,
		},
		{
			name:    "cpi test delete error noPod",
			podInfo: podWithErrorAnno,
			wantErr: false,
		},
		{
			name:    "cpi test delete no annotation pod",
			podInfo: podWithNoAnno,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newCpiService("Cpi")
			service.AddPod(tt.podInfo)
			if err := service.DeletePod(tt.podInfo); (err != nil) != tt.wantErr {
				t.Errorf("CpiService.AddPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCpiServiceCollectData tests CollectData function
func TestCpiServiceCollectData(t *testing.T) {
	onlinePod := &typedef.PodInfo{
		Name:      "onlinePod",
		UID:       "onlinePod",
		Hierarchy: cgroup.Hierarchy{},
		Annotations: map[string]string{
			constant.CpiAnnotationKey: "online",
		},
	}
	offlinePod := &typedef.PodInfo{
		Name:      "offlinePod",
		UID:       "offlinePod",
		Hierarchy: cgroup.Hierarchy{},
		Annotations: map[string]string{
			constant.CpiAnnotationKey: "offline",
		},
	}
	t.Run("cpu test collect data", func(t *testing.T) {
		service := newCpiService("Cpi")
		service.AddPod(onlinePod)
		service.AddPod(offlinePod)
		service.collectData(defaultSampleDur)
	})
}

// TestCpiServiceIdentifyAntagonists tests IdentifyAntagonists function
func TestCpiServiceIdentifyAntagonists(t *testing.T) {
	createFile := func(dir string, fileName string, content string) {
		if err := os.MkdirAll(dir, constant.DefaultDirMode); err != nil {
			t.Errorf("error creating temp dir: %v", err)
		}
		file, err := os.Create(path.Join(dir, fileName))
		file.Chmod(constant.DefaultFileMode)
		if err != nil {
			t.Errorf("error creating quota file: %v", err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Errorf("error writing to quota file: %v", err)
		}
		file.Close()
	}
	var fileName = "cpu.cfs_quota_us"
	tests := []struct {
		name         string
		onlineTasks  map[string]*podStatus
		offlineTasks map[string]*podStatus
		limitedIndex int
	}{
		{
			name: "without outlier",
			onlineTasks: map[string]*podStatus{
				"testOnline1": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "online",
					},
					isOnline: true,
					cpiSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 2.8, 2: 2.8, 3: 2.8, 4: 2.8, 5: 2.8, 6: 2.8, 7: 2.8, 8: 2.8, 9: 2.8, 10: 2.8},
					},
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 3, 2: 3, 3: 3, 4: 3, 5: 3, 6: 3, 7: 3, 8: 3, 9: 3, 10: 3},
					},
					cpiMean: 2.8,
					stdDev:  0,
					count:   11,
				},
			},
			limitedIndex: -1,
		},
		{
			name: "with out offlineTasks",
			onlineTasks: map[string]*podStatus{
				"testOnline1": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "online",
					},
					isOnline: true,
					cpiSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 2.8, 2: 2.8, 3: 2.8, 4: 2.8, 5: 2.8, 6: 2.8, 7: 2.8, 8: 2.8, 9: 2.8, 10: 2.8},
					},
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 3, 2: 3, 3: 3, 4: 3, 5: 3, 6: 3, 7: 3, 8: 3, 9: 3, 10: 3},
					},
					cpiMean: 2.8,
					stdDev:  0,
					count:   11,
				},
			},
			limitedIndex: -1,
		},
		{
			name: "No antagonist, but offline tasks exist.",
			onlineTasks: map[string]*podStatus{
				"testOnline1": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "online",
					},
					isOnline: true,
					cpiSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 2.1, 2: 2.1, 3: 2.1, 4: 2.1, 5: 2.1, 6: 2.1, 7: 2.1, 8: 2.1, 9: 2.1, 10: 2.1},
					},
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 3, 2: 3, 3: 3, 4: 3, 5: 3, 6: 3, 7: 3, 8: 3, 9: 3, 10: 3},
					},
					cpiMean: 2.0,
					stdDev:  0,
					count:   11,
				},
			},
			offlineTasks: map[string]*podStatus{
				"maxCpuUsagePod": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "offline",
					},
					isOnline: false,
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 6, 2: 6, 3: 6, 4: 6, 5: 6, 6: 6, 7: 6, 8: 6, 9: 6, 10: 6},
					},
				},
			},
			limitedIndex: 0,
		},
		{
			name: "Antagonist detected.",
			onlineTasks: map[string]*podStatus{
				"testOnline1": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "online",
					},
					isOnline: true,
					cpiSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 2.1, 2: 2.1, 3: 2.1, 4: 3.1, 5: 2.1, 6: 2.1, 7: 2.1, 8: 2.1, 9: 2.1, 10: 2.1},
					},
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 3, 2: 3, 3: 3, 4: 3, 5: 3, 6: 3, 7: 3, 8: 3, 9: 3, 10: 3},
					},
					cpiMean: 2.0,
					stdDev:  0,
					count:   11,
				},
			},
			offlineTasks: map[string]*podStatus{
				"maxCpuUsagePod": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "offline1",
					},
					isOnline: false,
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 6, 2: 6, 3: 6, 4: 6, 5: 6, 6: 6, 7: 6, 8: 6, 9: 6, 10: 6},
					},
				},
				"antagonistPod": {
					Hierarchy: &cgroup.Hierarchy{
						MountPoint: "/",
						Path:       "offline2",
					},
					isOnline: false,
					cpuUsageSeries: &dataSeries{
						timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						data:     map[int64]float64{1: 1, 2: 1, 3: 1, 4: 10, 5: 1, 6: 1, 7: 1, 8: 1, 9: 1, 10: 1},
					},
				},
			},
			limitedIndex: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &CpiService{
				onlineTasks:  tt.onlineTasks,
				offlineTasks: tt.offlineTasks,
			}
			now := time.Now()
			for podUID, onlinePod := range tt.onlineTasks {
				newOnline, _ := newPodStatus(&typedef.PodInfo{Hierarchy: *onlinePod.Hierarchy}, true, podUID)

				for _, minutes := range onlinePod.cpiSeries.timeline {
					newOnline.cpiSeries.add(onlinePod.cpiSeries.data[minutes], now.Add(time.Duration(minutes-10)*time.Minute).UnixNano())
				}

				for _, minutes := range onlinePod.cpuUsageSeries.timeline {
					newOnline.cpuUsageSeries.add(onlinePod.cpuUsageSeries.data[minutes], now.Add(time.Duration(minutes-10)*time.Minute).UnixNano())
				}

				newOnline.cpiMean = onlinePod.cpiMean
				newOnline.stdDev = onlinePod.stdDev
				newOnline.count = onlinePod.count

				service.onlineTasks[podUID] = newOnline
			}

			for podUID, offlinePod := range tt.offlineTasks {
				createFile(path.Join("/cpu", offlinePod.MountPoint, offlinePod.Path), fileName, "10000")
				defer os.RemoveAll("/cpu")
				newOffline := &podStatus{
					isOnline:       false,
					Hierarchy:      offlinePod.Hierarchy,
					cpuUsageSeries: newDataSeries(),
				}
				for _, minutes := range offlinePod.cpuUsageSeries.timeline {
					newOffline.cpuUsageSeries.add(offlinePod.cpuUsageSeries.data[minutes], now.Add(time.Duration(minutes-10)*time.Minute).UnixNano())
				}
				service.offlineTasks[podUID] = newOffline
			}

			service.identifyAntagonists(time.Minute*10, time.Minute*5, time.Second*2, defaultAntagonistMetric, defaultLimitQuota)

			if tt.limitedIndex == -1 {
				return
			}
			var podUid string
			if tt.limitedIndex == 0 {
				podUid = "maxCpuUsagePod"
			} else {
				podUid = "antagonistPod"
			}
			limitedPod := service.offlineTasks[podUid]
			quota := limitedPod.GetCgroupAttr(cpuQuotaKey).Value
			if quota != defaultLimitQuota {
				t.Errorf("Pod has not been limited as expected")
			}
		})
	}
}
