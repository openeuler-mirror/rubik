// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li, Danni Xia
// Create: 2022-05-16
// Description: offline pod cache limit function

// Package cachelimit is for cache limiting
package cachelimit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"isula.org/rubik/pkg/checkpoint"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/try"
	"isula.org/rubik/pkg/typedef"

	"github.com/stretchr/testify/assert"
)

var podInfo = typedef.PodInfo{
	CgroupPath:      "kubepods/podaaa",
	Offline:         true,
	CacheLimitLevel: "dynamic",
}

func initCpm() {
	podID := "podabc"
	cpm = &checkpoint.Manager{
		Checkpoint: &checkpoint.Checkpoint{
			Pods: map[string]*typedef.PodInfo{
				podID: &podInfo,
			},
		},
	}
}

// TestLevelValid testcase
func TestLevelValid(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TC-normal cache limit level",
			args: args{level: lowLevel},
			want: true,
		},
		{
			name: "TC-abnormal cache limit level",
			args: args{level: "abnormal level"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := levelValid(tt.args.level); got != tt.want {
				t.Errorf("levelValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewCacheLimitPodInfo test NewCacheLimitPodInfo
func TestSyncLevel(t *testing.T) {
	podInfo := typedef.PodInfo{
		CgroupPath:      "kubepods/podaaa",
		Offline:         true,
		CacheLimitLevel: "invalid",
	}
	err := SyncLevel(&podInfo)
	assert.Equal(t, true, err != nil)

	podInfo.CacheLimitLevel = lowLevel
	err = SyncLevel(&podInfo)
	assert.NoError(t, err)
	assert.Equal(t, podInfo.CacheLimitLevel, lowLevel)

	defaultLimitMode = staticMode
	podInfo.CacheLimitLevel = ""
	err = SyncLevel(&podInfo)
	assert.NoError(t, err)
	assert.Equal(t, podInfo.CacheLimitLevel, maxLevel)

	defaultLimitMode = dynamicMode
	podInfo.CacheLimitLevel = ""
	err = SyncLevel(&podInfo)
	assert.NoError(t, err)
	assert.Equal(t, podInfo.CacheLimitLevel, dynamicLevel)
}

// TestWriteTasksToResctrl test writeTasksToResctrl
func TestWriteTasksToResctrl(t *testing.T) {
	initCpm()
	err := SyncLevel(&podInfo)
	assert.NoError(t, err)

	testDir := try.GenTestDir().String()
	config.CgroupRoot = testDir

	pid, procsFile, container := "12345", "cgroup.procs", "container1"
	podCPUCgroupPath := filepath.Join(testDir, "cpu", podInfo.CgroupPath)
	try.MkdirAll(filepath.Join(podCPUCgroupPath, container), constant.DefaultDirMode)
	err = writeTasksToResctrl(&podInfo, testDir)
	// pod cgroup.procs not exist, return error
	assert.Equal(t, true, err != nil)
	_, err = os.Create(filepath.Join(podCPUCgroupPath, procsFile))
	assert.NoError(t, err)
	try.WriteFile(filepath.Join(podCPUCgroupPath, container, procsFile), []byte(pid), constant.DefaultFileMode)

	err = writeTasksToResctrl(&podInfo, testDir)
	// resctrl tasks file not exist, return error
	assert.Equal(t, true, err != nil)

	resctrlSubDir, taskFile := dirPrefix+podInfo.CacheLimitLevel, "tasks"
	try.MkdirAll(filepath.Join(testDir, resctrlSubDir), constant.DefaultDirMode)
	err = writeTasksToResctrl(&podInfo, testDir)
	// write success
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(filepath.Join(testDir, resctrlSubDir, taskFile))
	assert.NoError(t, err)
	assert.Equal(t, pid, strings.TrimSpace(string(bytes)))

	// container pid already written
	err = writeTasksToResctrl(&podInfo, testDir)
	assert.NoError(t, err)

	config.CgroupRoot = constant.DefaultCgroupRoot
}

// TestSetCacheLimit test SetCacheLimit
func TestSetCacheLimit(t *testing.T) {
	initCpm()
	err := SetCacheLimit(&podInfo)
	assert.NoError(t, err)
}

// TestSyncCacheLimit test syncCacheLimit
func TestSyncCacheLimit(t *testing.T) {
	initCpm()
	syncCacheLimit()
}
