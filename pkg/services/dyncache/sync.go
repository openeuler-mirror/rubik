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
// Description: This file is used for cache limit sync setting

// Package dyncache is the service used for cache limit setting
package dyncache

import (
	"fmt"
	"path/filepath"
	"strings"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	levelLow     = "low"
	levelMiddle  = "middle"
	levelHigh    = "high"
	levelMax     = "max"
	levelDynamic = "dynamic"

	resctrlDirPrefix = "rubik_"
	schemataFileName = "schemata"
)

var validLevel = map[string]bool{
	levelLow:     true,
	levelMiddle:  true,
	levelHigh:    true,
	levelMax:     true,
	levelDynamic: true,
}

// SyncCacheLimit will continuously set cache limit with corresponding offline pods
func (c *DynCache) SyncCacheLimit() {
	for _, p := range c.listOfflinePods() {
		if err := c.syncLevel(p); err != nil {
			log.Errorf("sync cache limit level err: %v", err)
			continue
		}
		if err := c.writeTasksToResctrl(p); err != nil {
			log.Errorf("set cache limit for pod %v err: %v", p.UID, err)
			continue
		}
	}
}

// writeTasksToResctrl will write tasks running in containers into resctrl group
func (c *DynCache) writeTasksToResctrl(pod *typedef.PodInfo) error {
	if !util.PathExist(cgroup.AbsoluteCgroupPath("cpu", pod.CgroupPath, "")) {
		// just return since pod maybe deleted
		return nil
	}
	var taskList []string
	cgroupKey := &cgroup.Key{SubSys: "cpu", FileName: "cgroup.procs"}
	for _, container := range pod.IDContainersMap {
		key := container.GetCgroupAttr(cgroupKey)
		if key.Err != nil {
			return key.Err
		}
		taskList = append(taskList, strings.Split(key.Value, "\n")...)
	}
	if len(taskList) == 0 {
		return nil
	}

	resctrlTaskFile := filepath.Join(c.config.DefaultResctrlDir,
		resctrlDirPrefix+pod.Annotations[constant.CacheLimitAnnotationKey], "tasks")
	for _, task := range taskList {
		if err := util.WriteFile(resctrlTaskFile, task); err != nil {
			if strings.Contains(err.Error(), "no such process") {
				log.Errorf("pod %s task %s not exist", pod.UID, task)
				continue
			}
			return fmt.Errorf("add task %v to file %v error: %v", task, resctrlTaskFile, err)
		}
	}

	return nil
}

// syncLevel sync cache limit level
func (c *DynCache) syncLevel(pod *typedef.PodInfo) error {
	if pod.Annotations[constant.CacheLimitAnnotationKey] == "" {
		if c.config.DefaultLimitMode == modeStatic {
			pod.Annotations[constant.CacheLimitAnnotationKey] = levelMax
		} else {
			pod.Annotations[constant.CacheLimitAnnotationKey] = levelDynamic
		}
	}

	level := pod.Annotations[constant.CacheLimitAnnotationKey]
	if isValid, ok := validLevel[level]; !ok || !isValid {
		return fmt.Errorf("invalid cache limit level %v for pod: %v", level, pod.UID)
	}
	return nil
}

func (c *DynCache) listOfflinePods() map[string]*typedef.PodInfo {
	offlineValue := "true"
	return c.Viewer.ListPodsWithOptions(func(pi *typedef.PodInfo) bool {
		return pi.Annotations[constant.PriorityAnnotationKey] == offlineValue
	})
}
