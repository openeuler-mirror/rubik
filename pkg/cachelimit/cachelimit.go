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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
	"isula.org/rubik/pkg/util"
)

const (
	resctrlDir = "/sys/fs/resctrl"
	noProErr   = "no such process"
)

// SyncLevel sync cache limit level
func SyncLevel(pi *typedef.PodInfo) error {
	level := pi.CacheLimitLevel
	if level == "" {
		if defaultLimitMode == staticMode {
			pi.CacheLimitLevel = maxLevel
		} else {
			pi.CacheLimitLevel = dynamicLevel
		}
	}
	if !levelValid(pi.CacheLimitLevel) {
		return errors.Errorf("invalid cache limit level %v for pod: %v", level, pi.UID)
	}
	return nil
}

// syncCacheLimit sync cache limit for offline pods, as new processes may generate during pod running,
// they should be moved to resctrl directory
func syncCacheLimit() {
	offlinePods := cpm.ListOfflinePods()
	for _, p := range offlinePods {
		if err := SyncLevel(p); err != nil {
			log.Errorf("sync cache limit level err: %v", err)
			continue
		}
		if err := writeTasksToResctrl(p, resctrlDir); err != nil {
			log.Errorf("set cache limit for pod %v err: %v", p.UID, err)
		}
	}
}

// SetCacheLimit set cache limit for offline pods
func SetCacheLimit(pi *typedef.PodInfo) error {
	log.Logf("setting cache limit level=%v for pod %s", pi.CacheLimitLevel, pi.UID)

	return writeTasksToResctrl(pi, resctrlDir)
}

func writeTasksToResctrl(pi *typedef.PodInfo, resctrlRoot string) error {
	taskRootPath := filepath.Join(config.CgroupRoot, "cpu", pi.CgroupPath)
	if !util.PathExist(taskRootPath) {
		log.Infof("path %v not exist, maybe pod %v is deleted", taskRootPath, pi.UID)
		return nil
	}

	tasks, _, err := getTasks(pi, taskRootPath)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	resctrlTaskFile := filepath.Join(resctrlRoot, dirPrefix+pi.CacheLimitLevel, "tasks")
	for _, task := range tasks {
		if err := ioutil.WriteFile(resctrlTaskFile, []byte(task), constant.DefaultFileMode); err != nil {
			if strings.Contains(err.Error(), noProErr) {
				log.Errorf("pod %s task %s not exist", pi.UID, task)
				continue
			}
			return errors.Errorf("add task %v to file %v error: %v", task, resctrlTaskFile, err)
		}
	}

	return nil
}

func getTasks(pi *typedef.PodInfo, taskRootPath string) ([]string, []string, error) {
	file := "cgroup.procs"
	var taskList, containers []string
	err := filepath.Walk(taskRootPath, func(path string, f os.FileInfo, err error) error {
		if f != nil && f.IsDir() {
			containerID := filepath.Base(f.Name())
			if cpm.ContainerExist(types.UID(pi.UID), containerID) {
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
