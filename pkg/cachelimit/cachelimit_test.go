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
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"isula.org/rubik/pkg/constant"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/api"
)

func initCacheLimitPod() {
	cacheLimitPods.Lock()
	defer cacheLimitPods.Unlock()
	cacheLimitPods.pods = make(map[string]*PodInfo, 0)
}

func (podList *podMap) getPodInfo(podID string) *PodInfo {
	podList.Lock()
	defer podList.Unlock()
	if podList.pods[podID] == nil {
		return nil
	}
	pod := PodInfo{
		podID:           podList.pods[podID].podID,
		cgroupPath:      podList.pods[podID].cgroupPath,
		cacheLimitLevel: podList.pods[podID].cacheLimitLevel,
	}
	containers := make(map[string]struct{}, 0)
	for i, c := range podList.pods[podID].containers {
		containers[i] = c
	}
	pod.containers = containers

	return &pod
}

// TestCheckCacheLimitLevel testcase
func TestCheckCacheLimitLevel(t *testing.T) {
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
func TestNewCacheLimitPodInfo(t *testing.T) {
	ctx := context.Background()
	podID := "podabc"
	podReq := api.PodQoS{
		CgroupPath:      "kubepods/podaaa",
		QosLevel:        -1,
		CacheLimitLevel: "invalid",
	}
	_, err := NewCacheLimitPodInfo(ctx, podID, podReq)
	assert.Equal(t, true, err != nil)

	podReq.CacheLimitLevel = lowLevel
	pod, err := NewCacheLimitPodInfo(ctx, podID, podReq)
	assert.NoError(t, err)
	assert.Equal(t, pod.podID, podID)
	assert.Equal(t, pod.cacheLimitLevel, podReq.CacheLimitLevel)
	assert.Equal(t, pod.cgroupPath, podReq.CgroupPath)

	defaultLimitMode = staticMode
	podReq.CacheLimitLevel = ""
	pod, err = NewCacheLimitPodInfo(ctx, podID, podReq)
	assert.NoError(t, err)
	assert.Equal(t, pod.cacheLimitLevel, maxLevel)

	defaultLimitMode = dynamicMode
	podReq.CacheLimitLevel = ""
	pod, err = NewCacheLimitPodInfo(ctx, podID, podReq)
	assert.NoError(t, err)
	assert.Equal(t, pod.cacheLimitLevel, dynamicLevel)
}

func newTestPodInfo(podID string) api.PodQoS {
	return api.PodQoS{
		CgroupPath:      "kubepods/" + podID,
		QosLevel:        -1,
		CacheLimitLevel: "low",
	}
}

// TestPodListAddAndDel test Add and Del
func TestPodListAddAndDel(t *testing.T) {
	initCacheLimitPod()
	podID := "podabc"
	pod, err := NewCacheLimitPodInfo(context.Background(), podID, newTestPodInfo(podID))
	assert.NoError(t, err)

	cacheLimitPods.Add(pod)
	assert.Equal(t, cacheLimitPods.getPodInfo(podID).podID, podID)
	cacheLimitPods.Del(podID)
	assert.Equal(t, true, cacheLimitPods.getPodInfo(podID) == nil)
}

// TestClone test clone
func TestClone(t *testing.T) {
	initCacheLimitPod()
	podID1 := "podabc"
	pod1, err := NewCacheLimitPodInfo(context.Background(), podID1, newTestPodInfo(podID1))
	pod1.containers["con1"] = struct{}{}
	assert.NoError(t, err)
	podID2 := "podabc2"
	pod2, err := NewCacheLimitPodInfo(context.Background(), podID2, newTestPodInfo(podID2))

	cacheLimitPods.Add(pod1)
	cacheLimitPods.Add(pod2)
	podList := cacheLimitPods.clone()
	podNum := 2
	assert.Equal(t, podNum, len(podList))
	cacheLimitPods.pods[podID1].containers["con2"] = struct{}{}
	cacheLimitPods.Del(podID2)
	for _, p := range podList {
		if p.podID == podID1 {
			assert.Equal(t, 1, len(p.containers))
		}
	}
}

// TestAddContainer test addContainer
func TestAddContainer(t *testing.T) {
	initCacheLimitPod()
	podID := "podabc"
	containerID := "containerabc"
	cacheLimitPods.addContainer(podID, containerID)
	assert.Equal(t, true, cacheLimitPods.getPodInfo(podID) == nil)
	assert.Equal(t, false, cacheLimitPods.containerExist(podID, containerID))

	pod, err := NewCacheLimitPodInfo(context.Background(), podID, newTestPodInfo(podID))
	assert.NoError(t, err)
	cacheLimitPods.Add(pod)
	cacheLimitPods.addContainer(podID, containerID)
	assert.Equal(t, true, cacheLimitPods.getPodInfo(podID).containers != nil)
	assert.Equal(t, true, cacheLimitPods.containerExist(podID, containerID))
}

// TestAddOnlinePod test AddOnlinePod
func TestAddOnlinePod(t *testing.T) {
	podID := "podabc"
	cgroupPath := "kubepods/podabc"
	AddOnlinePod(podID, cgroupPath)
	assert.Equal(t, podID, onlinePods.getPodInfo(podID).podID)
	assert.Equal(t, cgroupPath, onlinePods.getPodInfo(podID).cgroupPath)
}

// TestWriteTasksToResctrl test writeTasksToResctrl
func TestWriteTasksToResctrl(t *testing.T) {
	initCacheLimitPod()
	podID := "podabc"
	pod, err := NewCacheLimitPodInfo(context.Background(), podID, newTestPodInfo(podID))
	cacheLimitPods.Add(pod)
	assert.NoError(t, err)

	testDir := filepath.Join(constant.TmpTestDir, "cltest")
	err = os.MkdirAll(testDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer func() {
		err = os.RemoveAll(testDir)
		assert.NoError(t, err)
	}()

	pid, procsFile, container := "12345", "cgroup.procs", "container1"
	podCPUCgroupPath := filepath.Join(testDir, "cpu", pod.cgroupPath)
	err = os.MkdirAll(filepath.Join(podCPUCgroupPath, container), constant.DefaultDirMode)
	assert.NoError(t, err)
	err = pod.writeTasksToResctrl(testDir, testDir)
	// pod cgroup.procs not exist, return error
	assert.Equal(t, true, err != nil)
	_, err = os.Create(filepath.Join(podCPUCgroupPath, procsFile))
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(podCPUCgroupPath, container, procsFile), []byte(pid),
		constant.DefaultFileMode)
	assert.NoError(t, err)

	err = pod.writeTasksToResctrl(testDir, testDir)
	// resctrl tasks file not exist, return error
	assert.Equal(t, true, err != nil)

	resctrlSubDir, taskFile := dirPrefix+pod.cacheLimitLevel, "tasks"
	err = os.MkdirAll(filepath.Join(testDir, resctrlSubDir), constant.DefaultDirMode)
	assert.NoError(t, err)
	err = pod.writeTasksToResctrl(testDir, testDir)
	// write success
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(filepath.Join(testDir, resctrlSubDir, taskFile))
	assert.NoError(t, err)
	assert.Equal(t, pid, strings.TrimSpace(string(bytes)))

	// container pid already written
	err = pod.writeTasksToResctrl(testDir, testDir)
	assert.NoError(t, err)
}

// TestSetCacheLimit test SetCacheLimit
func TestSetCacheLimit(t *testing.T) {
	podID := "podabc"
	pod, err := NewCacheLimitPodInfo(context.Background(), podID, newTestPodInfo(podID))
	assert.NoError(t, err)
	err = pod.SetCacheLimit()
	assert.NoError(t, err)
}

// TestSyncCacheLimit test syncCacheLimit
func TestSyncCacheLimit(t *testing.T) {
	initCacheLimitPod()
	podID := "podabc"
	pod, err := NewCacheLimitPodInfo(context.Background(), podID, newTestPodInfo(podID))
	cacheLimitPods.Add(pod)
	assert.NoError(t, err)
	syncCacheLimit()
}
