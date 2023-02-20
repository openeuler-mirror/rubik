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
// Description: This file is used for testing data.go

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

// TestNodeDataGetLastCPUUtil tests getLastCPUUtil of NodeData
func TestNodeDataGetLastCPUUtil(t *testing.T) {
	// 1. empty CPU Utils
	d := &NodeData{}
	t.Run("TC1-empty CPU Util", func(t *testing.T) {
		util := float64(0.0)
		assert.Equal(t, util, d.getLastCPUUtil())
	})
	// 2. CPU Utils
	cpuUtil20 := 20
	d = &NodeData{cpuUtils: []cpuUtil{{
		util: float64(cpuUtil20),
	}}}
	t.Run("TC2-CPU Util is 20", func(t *testing.T) {
		util := float64(20.0)
		assert.Equal(t, util, d.getLastCPUUtil())
	})
}

// TestNodeDataRemoveContainer tests removeContainer of NodeData
func TestNodeDataRemoveContainer(t *testing.T) {
	var nodeDataRemoveContainerTests = []struct {
		data *NodeData
		name string
		id   string
		num  int
	}{
		{
			name: "TC1-no container exists",
			id:   "",
			num:  1,
			data: &NodeData{
				containers: map[string]*CPUQuota{
					"testCon1": {},
				},
			},
		},
		{
			name: "TC2-the container path does not exist",
			id:   "testCon3",
			num:  1,
			data: &NodeData{
				containers: map[string]*CPUQuota{
					"testCon1": {},
					"testCon3": {
						ContainerInfo: &typedef.ContainerInfo{
							ID:         "testCon3",
							CgroupPath: "kubepods/testPod3/testCon3",
						},
					},
				},
			},
		},
		{
			name: "TC3-delete a container normally",
			id:   containerInfos[0].ID,
			num:  0,
			data: &NodeData{
				containers: map[string]*CPUQuota{
					containerInfos[0].ID: {
						ContainerInfo: containerInfos[0].DeepCopy(),
						period:        100000,
					},
				},
			},
		},
		{
			name: "TC4-write an invalid value",
			id:   containerInfos[0].ID,
			num:  1,
			data: &NodeData{
				containers: map[string]*CPUQuota{
					containerInfos[0].ID: {
						ContainerInfo: containerInfos[0].DeepCopy(),
						period:        1000000000000000000,
					},
				},
			},
		},
	}
	cis := []*typedef.ContainerInfo{containerInfos[0].DeepCopy()}
	mkCgDirs(cis)
	defer rmCgDirs(cis)

	for _, tt := range nodeDataRemoveContainerTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.data.removeContainer(tt.id)
			assert.Equal(t, tt.num, len(tt.data.containers))
		})
	}
	us, err := cis[0].GetCgroupAttr(cpuQuotaKey).Int64()
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, int64(cis[0].LimitResources[typedef.ResourceCPU]*100000), us)
}

// TestQuotaTurboUpdateCPUUtils tests updateCPUUtils of QuotaTurbo and NewProcStat
func TestQuotaTurboUpdateCPUUtils(t *testing.T) {
	data := NewNodeData()
	// 1. obtain the cpu usage for the first time
	if err := data.updateCPUUtils(); err != nil {
		assert.NoError(t, err)
	}
	num1 := 1
	assert.Equal(t, num1, len(data.cpuUtils))
	// 2. obtain the cpu usage for the second time
	if err := data.updateCPUUtils(); err != nil {
		assert.NoError(t, err)
	}
	num2 := 2
	assert.Equal(t, num2, len(data.cpuUtils))
	// 3. obtain the cpu usage after 1 minute
	var minuteTimeDelta int64 = 60000000001
	data.cpuUtils[0].timestamp -= minuteTimeDelta
	if err := data.updateCPUUtils(); err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, num2, len(data.cpuUtils))
}

// TestIsAdjustmentAllowed tests isAdjustmentAllowed
func TestIsAdjustmentAllowed(t *testing.T) {
	cis := []*typedef.ContainerInfo{containerInfos[0].DeepCopy(), containerInfos[3].DeepCopy()}
	mkCgDirs(cis)
	defer rmCgDirs(cis)

	tests := []struct {
		ci   *typedef.ContainerInfo
		name string
		want bool
	}{
		{
			name: "TC1-allow adjustment",
			ci:   cis[0],
			want: true,
		},
		{
			name: "TC2-cgroup path is not existed",
			ci:   containerInfos[1],
			want: false,
		},
		{
			name: "TC3-cpulimit = 0",
			ci:   cis[1],
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, isAdjustmentAllowed(tt.ci), tt.want)
		})
	}
}

// TestUpdateClusterContainers tests UpdateClusterContainers
func TestUpdateClusterContainers(t *testing.T) {
	cis := []*typedef.ContainerInfo{containerInfos[0]}
	mkCgDirs(cis)
	defer rmCgDirs(cis)

	qt := &QuotaTurbo{
		NodeData: &NodeData{
			containers: make(map[string]*CPUQuota, 0),
		},
	}

	// 1. add container
	qt.UpdateClusterContainers(map[string]*typedef.ContainerInfo{cis[0].ID: cis[0]})
	conNum1 := 1
	assert.Equal(t, conNum1, len(qt.containers))

	// 2. updated container
	qt.UpdateClusterContainers(map[string]*typedef.ContainerInfo{cis[0].ID: cis[0]})
	assert.Equal(t, conNum1, len(qt.containers))

	// 3. deleting a container that does not meet the conditions (cgroup path is not existed)
	// 4. delete a container whose checkpoint does not exist.
	qt.containers[containerInfos[2].ID] = &CPUQuota{ContainerInfo: containerInfos[2].DeepCopy()}
	conNum2 := 2
	assert.Equal(t, conNum2, len(qt.containers))
	qt.UpdateClusterContainers(map[string]*typedef.ContainerInfo{containerInfos[2].ID: containerInfos[2].DeepCopy()})
	conNum0 := 0
	assert.Equal(t, conNum0, len(qt.containers))
}

// mkCgDirs creates the cgroup folder for the container
func mkCgDirs(cc []*typedef.ContainerInfo) {
	for _, ci := range cc {
		dirPath := cgroup.AbsoluteCgroupPath("cpu", ci.CgroupPath, "")
		fmt.Println("path : " + dirPath)
		try.MkdirAll(dirPath, constant.DefaultDirMode)
	}
}

// rmCgDirs deletes the cgroup folder of the container
func rmCgDirs(cc []*typedef.ContainerInfo) {
	for _, ci := range cc {
		dirPath := cgroup.AbsoluteCgroupPath("cpu", ci.CgroupPath, "")
		try.RemoveAll(dirPath)
		try.RemoveAll(path.Dir(dirPath))
	}
}
