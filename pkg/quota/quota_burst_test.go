// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Yanting Song
// Create: 2022-07-19
// Description: This file is used for test quota burst

package quota

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/typedef"
)

const (
	cfsBurstUs = "cpu.cfs_burst_us"
	cpuSubsys = "cpu"
)

var cis = []*typedef.ContainerInfo{
	{
		Name:       "FooCon",
		ID:         "testCon1",
		PodID:      "testPod1",
		CgroupRoot: constant.TmpTestDir,
		CgroupAddr: "kubepods/testPod1/testCon1",
	},
	{
		Name:       "BarCon",
		ID:         "testCon2",
		PodID:      "testPod2",
		CgroupRoot: constant.TmpTestDir,
		CgroupAddr: "kubepods/testPod2/testCon2",
	},
	{
		Name:       "BiuCon",
		ID:         "testCon3",
		PodID:      "testPod3",
		CgroupRoot: constant.TmpTestDir,
		CgroupAddr: "kubepods/testPod3/testCon3",
	},
	{
		Name:       "NotExist",
		ID:         "testCon4",
		PodID:      "testPod4",
		CgroupRoot: constant.TmpTestDir,
		CgroupAddr: "kubepods/testPod4/testCon4",
	},
}

var pis = []*typedef.PodInfo{
	// valid QuotaBurstQuota value
	{
		Name: "FooPod",
		UID:  cis[0].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			cis[0].Name: cis[0],
		},
		QuotaBurst: 0,
	},
	// invalid QuotaBurstQuota value
	{
		Name: "BarPod",
		UID:  cis[1].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			cis[1].Name: cis[1],
		},
		QuotaBurst: -1,
	},
	// valid QuotaBurstQuota value
	{
		Name: "BiuPod",
		UID:  cis[2].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			cis[2].Name: cis[2],
		},
		QuotaBurst: 1,
	},
}

var notExistPod = &typedef.PodInfo{
	Name: "NotExistPod",
	UID:  cis[3].PodID,
	Containers: map[string]*typedef.ContainerInfo{
		cis[3].Name: cis[3],
	},
	QuotaBurst: 0,
}

type updateQuotaBurstTestCase struct {
	oldPodInfo *typedef.PodInfo
	newPodInfo *typedef.PodInfo
	name       string
	wantValue  string
}

type podQuotaBurstTestCase struct {
	podInfo   *typedef.PodInfo
	name      string
	wantValue string
}

func createCgroupPath(t *testing.T) error {
	for _, ci := range pis {
		for _, ctr := range ci.Containers {
			ctrAddr := filepath.Join(constant.TmpTestDir, cpuSubsys, ctr.CgroupAddr)
			if err := os.MkdirAll(ctrAddr, constant.DefaultDirMode); err != nil {
				return err
			}
			if _, err := os.Create(filepath.Join(ctrAddr, cfsBurstUs)); err != nil {
				return err
			}
		}
	}
	return nil
}

// TestSetPodsQuotaBurst tests auto check pod's quota burst
func TestSetPodsQuotaBurst(t *testing.T) {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer os.RemoveAll(constant.TmpTestDir)

	pods := make(map[string]*typedef.PodInfo, len(pis))
	for _, pi := range pis {
		pods[pi.Name] = pi
	}
	if err := createCgroupPath(t); err != nil {
		t.Errorf("createCgroupPath got %v ", err)
	}

	SetPodsQuotaBurst(pods)

	for i, pi := range pis {
		ctrAddr := filepath.Join(constant.TmpTestDir, cpuSubsys, cis[i].CgroupAddr)
		var quotaBurst []byte
		if quotaBurst, err = ioutil.ReadFile(filepath.Join(ctrAddr, cfsBurstUs)); err != nil {
			t.Errorf("readFile got %v ", err)
		}
		var expected string
		if pi.QuotaBurst == constant.InvalidBurst {
			expected = ""
		} else {
			expected = strconv.Itoa(int(pi.QuotaBurst))
		}
		assert.Equal(t, expected, string(quotaBurst))
	}
}

// TestUpdateCtrQuotaBurst tests update pod's quota burst
func TestUpdateCtrQuotaBurst(t *testing.T) {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer os.RemoveAll(constant.TmpTestDir)

	if err := createCgroupPath(t); err != nil {
		t.Errorf("createCgroupPath got %v ", err)
	}
	for _, tt := range []updateQuotaBurstTestCase{
		{
			name:       "TC1-update quota burst with old podinfo is nil",
			oldPodInfo: nil,
			newPodInfo: pis[0],
			wantValue:  "",
		},
		{
			name:       "TC2-update quota burst without change",
			oldPodInfo: pis[0],
			newPodInfo: pis[0],
			wantValue:  "",
		},
		{
			name:       "TC3-update quota burst",
			oldPodInfo: pis[1],
			newPodInfo: pis[0],
			wantValue:  "0",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			UpdatePodQuotaBurst(tt.oldPodInfo, tt.newPodInfo)
			for _, ctr := range tt.newPodInfo.Containers {
				ctrAddr := filepath.Join(constant.TmpTestDir, cpuSubsys, ctr.CgroupAddr)
				var quotaBurst []byte
				if quotaBurst, err = ioutil.ReadFile(filepath.Join(ctrAddr, cfsBurstUs)); err != nil {
					t.Errorf("readFile got %v ", err)
				}
				assert.Equal(t, tt.wantValue, string(quotaBurst))
			}
		})
	}
}

// TestSetPodQuotaBurst tests set quota burst of pod
func TestSetPodQuotaBurst(t *testing.T) {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer os.RemoveAll(constant.TmpTestDir)
	if err := createCgroupPath(t); err != nil {
		t.Errorf("createCgroupPath got %v ", err)
	}
	err = setCtrQuotaBurst([]byte("0"), notExistPod.Containers["NotExist"])
	assert.Contains(t, err.Error(), "missing")

	for i, tt := range []podQuotaBurstTestCase{
		{
			name:      "TC1-set pod burst with valid quota value 0",
			podInfo:   pis[0],
			wantValue: "0",
		},
		{
			name:      "TC2-set pod burst with invalid quota value",
			podInfo:   pis[1],
			wantValue: "",
		},
		{
			name:      "TC3-set pod burst with nil podinfo",
			podInfo:   nil,
			wantValue: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			SetPodQuotaBurst(tt.podInfo)
			ctrAddr := filepath.Join(constant.TmpTestDir, cpuSubsys, cis[i].CgroupAddr)
			var quotaBurst []byte
			if quotaBurst, err = ioutil.ReadFile(filepath.Join(ctrAddr, cfsBurstUs)); err != nil {
				t.Errorf("readFile got %v ", err)
			}
			assert.Equal(t, tt.wantValue, string(quotaBurst))
		})
	}
}
