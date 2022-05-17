// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2022-01-18
// Description: offline pod cache limit function

// Package cachelimit is for cache limiting
package cachelimit

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/pkg/errors"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
)

const (
	resctrlDir = "/sys/fs/resctrl"
	noProErr   = "no such process"
)

var cacheLimitPods, onlinePods *podMap

type podMap struct {
	sync.RWMutex
	pods map[string]*PodInfo
}

// PodInfo describe pod information
type PodInfo struct {
	ctx             context.Context
	podID           string
	cgroupPath      string
	cacheLimitLevel string
	containers      map[string]struct{}
}

func init() {
	cacheLimitPods = newPodMap()
	onlinePods = newPodMap()
}

func newPodMap() *podMap {
	return &podMap{pods: make(map[string]*PodInfo, 0)}
}

// NewCacheLimitPodInfo return cache limit pod info
func NewCacheLimitPodInfo(ctx context.Context, podID string, podReq api.PodQoS) (*PodInfo, error) {
	level := podReq.CacheLimitLevel
	if level == "" {
		if defaultLimitMode == staticMode {
			level = maxLevel
		} else {
			level = dynamicLevel
		}
	}
	if !levelValid(level) {
		return nil, errors.Errorf("invalid cache limit level %v for pod: %v", level, podID)
	}

	return &PodInfo{
		ctx:             ctx,
		podID:           podID,
		cgroupPath:      podReq.CgroupPath,
		cacheLimitLevel: level,
		containers:      make(map[string]struct{}, 0),
	}, nil
}

// syncCacheLimit sync cache limit for offline pods, as new processes may generate during pod running,
// they should be moved to resctrl directory
func syncCacheLimit() {
	for _, p := range cacheLimitPods.clone() {
		if err := p.writeTasksToResctrl(config.CgroupRoot, resctrlDir); err != nil {
			log.Errorf("Set cache limit for pod %v err: %v", p.podID, err)
		}
	}
}

// Add add pod to pod list
func (pm *podMap) Add(pod *PodInfo) {
	pm.Lock()
	defer pm.Unlock()
	if _, ok := pm.pods[pod.podID]; !ok {
		pm.pods[pod.podID] = pod
	}
}

// Del remove pod from pod list
func (pm *podMap) Del(podID string) {
	pm.Lock()
	defer pm.Unlock()
	delete(pm.pods, podID)
}

func (pm *podMap) clone() []*PodInfo {
	var pods []*PodInfo
	pm.Lock()
	defer pm.Unlock()
	for _, pod := range pm.pods {
		pods = append(pods, pod.clone())
	}
	return pods
}

func (pm *podMap) addContainer(podID, containerID string) {
	pm.Lock()
	defer pm.Unlock()
	if _, ok := pm.pods[podID]; !ok {
		return
	}
	pm.pods[podID].containers[containerID] = struct{}{}
}

func (pm *podMap) containerExist(podID, containerID string) bool {
	pm.Lock()
	defer pm.Unlock()
	if _, ok := pm.pods[podID]; !ok {
		return false
	}
	if _, ok := pm.pods[podID].containers[containerID]; ok {
		return true
	}
	return false
}

func (pod *PodInfo) clone() *PodInfo {
	p := *pod
	p.containers = make(map[string]struct{}, len(pod.containers))
	for c := range pod.containers {
		p.containers[c] = struct{}{}
	}
	return &p
}

// SetCacheLimit set cache limit for offline pods
func (pod *PodInfo) SetCacheLimit() error {
	log.WithCtx(pod.ctx).Logf("Setting cache limit level=%v for pod %s", pod.cacheLimitLevel, pod.podID)
	cacheLimitPods.Add(pod)

	return pod.writeTasksToResctrl(config.CgroupRoot, resctrlDir)
}

func (pod *PodInfo) writeTasksToResctrl(cgroupRoot, resctrlRoot string) error {
	taskRootPath := filepath.Join(cgroupRoot, "cpu", pod.cgroupPath)
	if !util.PathExist(taskRootPath) {
		log.Infof("Path %v not exist, maybe pod %v is deleted", taskRootPath, pod.podID)
		cacheLimitPods.Del(pod.podID)
		return nil
	}

	tasks, containers, err := pod.getTasks(taskRootPath)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	resctrlTaskFile := filepath.Join(resctrlRoot, dirPrefix+pod.cacheLimitLevel, "tasks")
	success := true
	for _, task := range tasks {
		if err := ioutil.WriteFile(resctrlTaskFile, []byte(task), constant.DefaultFileMode); err != nil {
			success = false
			if strings.Contains(err.Error(), noProErr) {
				log.Errorf("pod %s task %s not exist", pod.podID, task)
				continue
			}
			return errors.Errorf("add task %v to file %v error: %v", task, resctrlTaskFile, err)
		}
	}
	if !success {
		return nil
	}

	for _, containerID := range containers {
		cacheLimitPods.addContainer(pod.podID, containerID)
	}

	return nil
}

func (pod *PodInfo) getTasks(taskRootPath string) ([]string, []string, error) {
	file := "cgroup.procs"
	var taskList, containers []string
	err := filepath.Walk(taskRootPath, func(path string, f os.FileInfo, err error) error {
		if f != nil && f.IsDir() {
			containerID := filepath.Base(f.Name())
			if cacheLimitPods.containerExist(pod.podID, containerID) {
				return nil
			}
			cgFilePath, err := securejoin.SecureJoin(path, file)
			if err != nil {
				return errors.Errorf("join path failed for %s and %s: %v", path, file, err)
			}
			tasks, err := ioutil.ReadFile(filepath.Clean(cgFilePath))
			if err != nil {
				return errors.Errorf("read task file %v err: %v", cgFilePath, err)
			}
			if strings.TrimSpace(string(tasks)) == "" {
				return nil
			}
			if containerID != filepath.Base(taskRootPath) {
				containers = append(containers, containerID)
			}
			taskList = append(taskList, strings.Split(strings.TrimSpace(string(tasks)), "\n")...)
		}
		return nil
	})

	return taskList, containers, err
}

// AddOnlinePod add pod to online pod list
func AddOnlinePod(podID, cgroupPath string) {
	onlinePods.Add(&PodInfo{
		podID:      podID,
		cgroupPath: cgroupPath,
		containers: make(map[string]struct{}, 0),
	})
}

func levelValid(level string) bool {
	switch level {
	case lowLevel:
	case middleLevel:
	case highLevel:
	case maxLevel:
	case dynamicLevel:
	default:
		return false
	}

	return true
}
