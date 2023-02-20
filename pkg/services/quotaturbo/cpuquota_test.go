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
// Description: This file is used for test cpu_quota

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
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

// TestSaveQuota tests SaveQuota of CPUQuota
func TestSaveQuota(t *testing.T) {
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
	cgroup.InitMountDir(constant.TmpTestDir)
	var (
		cq = &CPUQuota{
			ContainerInfo: &typedef.ContainerInfo{
				Name:       "Foo",
				ID:         "testCon1",
				CgroupPath: "kubepods/testPod1/testCon1",
			},
		}
		contPath       = cgroup.AbsoluteCgroupPath("cpu", cq.CgroupPath, "")
		podPeriodPath  = filepath.Join(filepath.Dir(contPath), cpuPeriodFile)
		podQuotaPath   = filepath.Join(filepath.Dir(contPath), cpuQuotaFile)
		contPeriodPath = filepath.Join(contPath, cpuPeriodFile)
		contQuotaPath  = filepath.Join(contPath, cpuQuotaFile)
	)

	try.RemoveAll(constant.TmpTestDir)
	defer try.RemoveAll(constant.TmpTestDir)

	assertValue := func(t *testing.T, paths []string, value string) {
		for _, p := range paths {
			data, err := util.ReadFile(p)
			assert.NoError(t, err)
			assert.Equal(t, value, strings.TrimSpace(string(data)))
		}
	}

	// case1: Only one pod or container exists at a time
	func() {
		try.MkdirAll(contPath, constant.DefaultDirMode)
		defer try.RemoveAll(path.Dir(contPath))
		// None of the paths exist
		cq.nextQuota = largerQuotaVal
		cq.curQuota = smallerQuotaVal
		assert.Error(t, cq.SaveQuota())
		cq.nextQuota = smallerQuotaVal
		cq.curQuota = largerQuotaVal
		assert.Error(t, cq.SaveQuota())

		// only Pod path existed
		cq.nextQuota = largerQuotaVal
		cq.curQuota = smallerQuotaVal
		try.WriteFile(podQuotaPath, unlimitedQuota)
		try.WriteFile(podPeriodPath, periodUs)
		try.RemoveAll(contQuotaPath)
		try.RemoveAll(contPeriodPath)
		assert.Error(t, cq.SaveQuota())

		// only container path existed
		cq.nextQuota = smallerQuotaVal
		cq.curQuota = largerQuotaVal
		try.RemoveAll(podQuotaPath)
		try.RemoveAll(podPeriodPath)
		try.WriteFile(contQuotaPath, largerQuota)
		try.WriteFile(contPeriodPath, periodUs)
		assert.Error(t, cq.SaveQuota())

	}()

	// case2: success
	func() {
		try.MkdirAll(contPath, constant.DefaultDirMode)
		defer try.RemoveAll(path.Dir(contPath))

		try.WriteFile(podQuotaPath, smallerQuota)
		try.WriteFile(podPeriodPath, periodUs)
		try.WriteFile(contQuotaPath, smallerQuota)
		try.WriteFile(contPeriodPath, periodUs)

		// delta > 0
		cq.nextQuota = largerQuotaVal
		cq.curQuota = smallerQuotaVal
		assert.NoError(t, cq.SaveQuota())
		assertValue(t, []string{podQuotaPath, contQuotaPath}, largerQuota)
		// delta < 0
		cq.nextQuota = smallerQuotaVal
		cq.curQuota = largerQuotaVal
		assert.NoError(t, cq.SaveQuota())
		assertValue(t, []string{podQuotaPath, contQuotaPath}, smallerQuota)
	}()

	cgroup.InitMountDir(constant.DefaultCgroupRoot)
}

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
	cgroup.InitMountDir(constant.TmpTestDir)

	var (
		ci = &typedef.ContainerInfo{
			Name:       "Foo",
			ID:         "testCon1",
			CgroupPath: "kubepods/testPod1/testCon1",
		}
		contPath       = cgroup.AbsoluteCgroupPath("cpu", ci.CgroupPath, "")
		contPeriodPath = filepath.Join(contPath, cpuPeriodFile)
		contQuotaPath  = filepath.Join(contPath, cpuQuotaFile)
		contUsagePath  = cgroup.AbsoluteCgroupPath("cpuacct", ci.CgroupPath, cpuUsageFile)
		contStatPath   = filepath.Join(contPath, cpuStatFile)
	)

	try.RemoveAll(constant.TmpTestDir)
	try.MkdirAll(contPath, constant.DefaultDirMode)
	try.MkdirAll(path.Dir(contUsagePath), constant.DefaultDirMode)
	defer try.RemoveAll(constant.TmpTestDir)

	// absent of period file
	try.RemoveAll(contPeriodPath)
	_, err := NewCPUQuota(ci)
	assert.Error(t, err, "should lacking of period file")
	try.WriteFile(contPeriodPath, period)

	// absent of throttle file
	try.RemoveAll(contStatPath)
	_, err = NewCPUQuota(ci)
	assert.Error(t, err, "should lacking of throttle file")
	try.WriteFile(contStatPath, validStat)

	// absent of quota file
	try.RemoveAll(contQuotaPath)
	_, err = NewCPUQuota(ci)
	assert.Error(t, err, "should lacking of quota file")
	try.WriteFile(contQuotaPath, quota)

	// absent of usage file
	try.RemoveAll(contUsagePath)
	_, err = NewCPUQuota(ci)
	assert.Error(t, err, "should lacking of usage file")
	try.WriteFile(contUsagePath, usage)

	cq, err := NewCPUQuota(ci)
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
	cgroup.InitMountDir(constant.DefaultCgroupRoot)
}

var containerInfos = []*typedef.ContainerInfo{
	{
		Name:             "Foo",
		ID:               "testCon1",
		CgroupPath:       "kubepods/testPod1/testCon1",
		LimitResources:   typedef.ResourceMap{typedef.ResourceCPU: 2},
		RequestResources: typedef.ResourceMap{typedef.ResourceCPU: 2},
	},
	{
		Name:             "Bar",
		ID:               "testCon2",
		CgroupPath:       "kubepods/testPod2/testCon2",
		LimitResources:   typedef.ResourceMap{typedef.ResourceCPU: 3},
		RequestResources: typedef.ResourceMap{typedef.ResourceCPU: 3},
	},
	{
		Name:             "Biu",
		ID:               "testCon3",
		CgroupPath:       "kubepods/testPod3/testCon3",
		LimitResources:   typedef.ResourceMap{typedef.ResourceCPU: 2},
		RequestResources: typedef.ResourceMap{typedef.ResourceCPU: 2},
	},
	{
		Name:             "Pah",
		ID:               "testCon4",
		CgroupPath:       "kubepods/testPod4/testCon4",
		LimitResources:   typedef.ResourceMap{typedef.ResourceCPU: 0},
		RequestResources: typedef.ResourceMap{typedef.ResourceCPU: 0},
	},
}
