// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2023-02-21
// Description: This file is testcases for dynamic cache limit level setting

// Package cachelimit is the service used for cache limit setting
package dynCache

import (
	"fmt"
	"math"
	"os"
	"testing"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

func TestCacheLimit_StartDynamic(t *testing.T) {
	if !perf.Support() {
		t.Skipf("%s only run on physical machine", t.Name())
	}
	try.InitTestCGRoot(constant.DefaultCgroupRoot)
	type fields struct {
		Config   *Config
		Attr     *Attr
		Name     string
		FakePods []*try.FakePod
	}
	tests := []struct {
		name     string
		fields   fields
		preHook  func(t *testing.T, c *DynCache, fakePods []*try.FakePod)
		postHook func(t *testing.T, c *DynCache, fakePods []*try.FakePod)
	}{
		{
			name: "TC-normal dynamic setting",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOnlinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
				},
				Config: genDefaultConfig(),
				Attr: &Attr{
					MaxMiss: 20,
					MinMiss: 10,
				},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				resctrlDir := try.GenTestDir().String()
				setMaskFile(t, resctrlDir, "3ff")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 4)
				c.Config.DefaultResctrlDir = resctrlDir
				c.Config.DefaultLimitMode = modeDynamic
				c.Config.PerfDuration = 10
				for _, pod := range fakePods {
					if pod.Annotations[constant.PriorityAnnotationKey] == "true" {
						pod.Annotations[constant.CacheLimitAnnotationKey] = "dynamic"
					}
				}
				manager := genPodManager(fakePods)
				c.Viewer = manager
			},
			postHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				for _, pod := range fakePods {
					pod.CleanPath()
				}
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
		{
			name: "TC-max and min miss both 0",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOnlinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
				},
				Config: genDefaultConfig(),
				Attr: &Attr{
					MaxMiss: 0,
					MinMiss: 0,
				},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				resctrlDir := try.GenTestDir().String()
				setMaskFile(t, resctrlDir, "3ff")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 4)
				c.Config.DefaultResctrlDir = resctrlDir
				c.Config.DefaultLimitMode = modeDynamic
				c.Config.PerfDuration = 10
				for _, pod := range fakePods {
					if pod.Annotations[constant.PriorityAnnotationKey] == "true" {
						pod.Annotations[constant.CacheLimitAnnotationKey] = "dynamic"
					}
				}
				manager := genPodManager(fakePods)
				c.Viewer = manager
			},
			postHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				for _, pod := range fakePods {
					pod.CleanPath()
				}
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
		{
			name: "TC-start dynamic with very high water line",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOnlinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
				},
				Config: genDefaultConfig(),
				Attr: &Attr{
					MaxMiss: math.MaxInt64,
					MinMiss: math.MaxInt64,
				},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				resctrlDir := try.GenTestDir().String()
				setMaskFile(t, resctrlDir, "3ff")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 4)
				c.Config.DefaultResctrlDir = resctrlDir
				c.Config.DefaultLimitMode = modeDynamic
				c.Config.PerfDuration = 10
				for _, pod := range fakePods {
					if pod.Annotations[constant.PriorityAnnotationKey] == "true" {
						pod.Annotations[constant.CacheLimitAnnotationKey] = "dynamic"
					}
				}
				manager := genPodManager(fakePods)
				c.Viewer = manager
			},
			postHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				for _, pod := range fakePods {
					pod.CleanPath()
				}
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
		{
			name: "TC-start dynamic with low min water line",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOnlinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
				},
				Config: genDefaultConfig(),
				Attr: &Attr{
					MaxMiss: math.MaxInt64,
					MinMiss: 0,
				},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				resctrlDir := try.GenTestDir().String()
				setMaskFile(t, resctrlDir, "3ff")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 4)
				c.Config.DefaultResctrlDir = resctrlDir
				c.Config.DefaultLimitMode = modeDynamic
				c.Config.PerfDuration = 10
				for _, pod := range fakePods {
					if pod.Annotations[constant.PriorityAnnotationKey] == "true" {
						pod.Annotations[constant.CacheLimitAnnotationKey] = "dynamic"
					}
				}
				manager := genPodManager(fakePods)
				c.Viewer = manager
			},
			postHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				for _, pod := range fakePods {
					pod.CleanPath()
				}
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
		{
			name: "TC-dynamic not exist",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOnlinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
					try.GenFakeOnlinePod(map[*cgroup.Key]string{
						{SubSys: "perf_event", FileName: "tasks"}: fmt.Sprintf("%d", os.Getegid()),
					}),
				},
				Config: genDefaultConfig(),
				Attr: &Attr{
					MaxMiss: 20,
					MinMiss: 10,
				},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				resctrlDir := try.GenTestDir().String()
				setMaskFile(t, resctrlDir, "3ff")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 4)
				c.Config.DefaultResctrlDir = resctrlDir
				c.Config.DefaultLimitMode = modeDynamic
				c.Config.PerfDuration = 10
				for _, pod := range fakePods {
					if pod.Annotations[constant.PriorityAnnotationKey] == "true" {
						pod.Annotations[constant.CacheLimitAnnotationKey] = "static"
					}
				}
				manager := genPodManager(fakePods)
				c.Viewer = manager
			},
			postHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				for _, pod := range fakePods {
					pod.CleanPath()
				}
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DynCache{
				Config: tt.fields.Config,
				Attr:   tt.fields.Attr,
				Name:   tt.fields.Name,
			}
			if tt.preHook != nil {
				tt.preHook(t, c, tt.fields.FakePods)
			}
			c.StartDynamic()
			if tt.postHook != nil {
				tt.postHook(t, c, tt.fields.FakePods)
			}
		})
	}
}
