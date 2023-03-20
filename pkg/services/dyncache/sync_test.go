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
// Description: This file is testcase for cache limit sync function

// Package dyncache is the service used for cache limit setting
package dyncache

import (
	"path/filepath"
	"testing"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/podmanager"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/test/try"
)

func genDefaultConfig() *Config {
	return &Config{
		DefaultLimitMode:    modeStatic,
		DefaultResctrlDir:   defaultResctrlDir,
		DefaultPidNameSpace: defaultPidNameSpace,
		AdjustInterval:      defaultAdInt,
		PerfDuration:        defaultPerfDur,
		L3Percent:           MultiLvlPercent{Low: defaultLowL3, Mid: defaultMidL3, High: defaultHighL3},
		MemBandPercent:      MultiLvlPercent{Low: defaultLowMB, Mid: defaultMidMB, High: defaultHighMB},
	}
}

func genPodManager(fakePods []*try.FakePod) *podmanager.PodManager {
	pm := &podmanager.PodManager{
		Pods: &podmanager.PodCache{
			Pods: make(map[string]*typedef.PodInfo, 0),
		},
	}
	for _, pod := range fakePods {
		pm.Pods.Pods[pod.UID] = pod.PodInfo
	}
	return pm
}

func cleanFakePods(fakePods []*try.FakePod) {
	for _, pod := range fakePods {
		pod.CleanPath().OrDie()
	}
}

// TestCacheLimit_SyncCacheLimit tests SyncCacheLimit of CacheLimit
func TestCacheLimit_SyncCacheLimit(t *testing.T) {
	resctrlDir := try.GenTestDir().String()
	defer try.RemoveAll(resctrlDir)
	try.InitTestCGRoot(try.TestRoot)
	defaultConfig := genDefaultConfig()
	defaultConfig.DefaultResctrlDir = resctrlDir
	type fields struct {
		Config   *Config
		Attr     *Attr
		Name     string
		FakePods []*try.FakePod
	}

	tests := []struct {
		name    string
		fields  fields
		preHook func(t *testing.T, c *DynCache, fakePods []*try.FakePod)
	}{
		{
			name: "TC1-normal case",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					pod.Annotations[constant.CacheLimitAnnotationKey] = "low"
					try.WriteFile(filepath.Join(defaultConfig.DefaultResctrlDir,
						resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks"), "")
				}
				c.Viewer = manager
			},
		},
		{
			name: "TC2-empty annotation with static mode config",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					try.WriteFile(filepath.Join(defaultConfig.DefaultResctrlDir,
						resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks"), "")
				}
				c.Viewer = manager
			},
		},
		{
			name: "TC3-empty annotation with dynamic mode config",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					try.WriteFile(filepath.Join(defaultConfig.DefaultResctrlDir,
						resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks"), "")
				}
				c.Viewer = manager
				c.config.DefaultLimitMode = levelDynamic
			},
		},
		{
			name: "TC4-invalid annotation",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					pod.Annotations[constant.CacheLimitAnnotationKey] = "invalid"
					try.WriteFile(filepath.Join(defaultConfig.DefaultResctrlDir,
						resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks"), "")
				}
				c.Viewer = manager
			},
		},
		{
			name: "TC5-pod just deleted",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					pod.Annotations[constant.CacheLimitAnnotationKey] = "low"
					try.RemoveAll(cgroup.AbsoluteCgroupPath("cpu", pod.CgroupPath, ""))
				}
				c.Viewer = manager
			},
		},
		{
			name: "TC6-pod without cgroup.procs",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "aaa"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					pod.Annotations[constant.CacheLimitAnnotationKey] = "low"
					try.WriteFile(filepath.Join(defaultConfig.DefaultResctrlDir,
						resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks"), "")
				}
				c.Viewer = manager
			},
		},
		{
			name: "TC7-pod without containers",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					pod.Annotations[constant.CacheLimitAnnotationKey] = "low"
					try.WriteFile(filepath.Join(defaultConfig.DefaultResctrlDir,
						resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks"), "")
				}
				c.Viewer = manager
			},
		},
		{
			name: "TC8-invalid resctrl group path",
			fields: fields{
				FakePods: []*try.FakePod{
					try.GenFakeOfflinePod(map[*cgroup.Key]string{
						{SubSys: "cpu", FileName: "cgroup.procs"}: "12345",
					}).WithContainers(1),
				},
				Config: defaultConfig,
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache, fakePods []*try.FakePod) {
				manager := genPodManager(fakePods)
				for _, pod := range manager.Pods.Pods {
					pod.Annotations[constant.CacheLimitAnnotationKey] = "low"
				}
				c.Viewer = manager
				c.config.DefaultResctrlDir = "/dev/null"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DynCache{
				config: tt.fields.Config,
				Attr:   tt.fields.Attr,
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
			}
			if tt.preHook != nil {
				tt.preHook(t, c, tt.fields.FakePods)
			}
			c.SyncCacheLimit()
			cleanFakePods(tt.fields.FakePods)
		})
	}
	try.RemoveAll(resctrlDir)
}
