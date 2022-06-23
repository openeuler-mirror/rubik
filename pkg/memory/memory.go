// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Song Yanting
// Create: 2022-06-10
// Description: memory setting for pods

package memory

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/pkg/errors"

	"isula.org/rubik/pkg/checkpoint"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
)

const (
	mlimit fileType = iota
	msoftLimit
)

const (
	dropCachesFilePath   = "/proc/sys/vm/drop_caches"
	memoryLimitFile      = "memory.limit_in_bytes"
	memorySoftLimitFile  = "memory.soft_limit_in_bytes"
	memoryUsageFile      = "memory.usage_in_bytes"
	memoryForceEmptyFile = "memory.force_empty"
	// maxSysMemLimit 9223372036854771712 is the default cgroup memory limit value
	maxSysMemLimit      = 9223372036854771712
	maxRetry            = 3
	relieveMaxCnt       = 5
	extraFreePercentage = 0.02
)

type fileType int

type memDriver interface {
	Run()
}

// MemoryManager manages memory reclaim works.
type MemoryManager struct {
	cpm           *checkpoint.Manager
	md            memDriver
	checkInterval int
	stop          chan struct{}
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(cpm *checkpoint.Manager, memConfig config.MemoryConfig) (*MemoryManager, error) {
	interval := memConfig.CheckInterval
	if err := validateInterval(interval); err != nil {
		return nil, err
	}
	log.Logf("new memory manager with interval:%d", interval)
	mm := MemoryManager{
		cpm:           cpm,
		checkInterval: interval,
		stop:          config.ShutdownChan,
	}
	switch memConfig.Strategy {
	case "dynlevel":
		mm.md = newDynLevel(&mm)
	case "none":
		log.Infof("strategy is set to none")
		return nil, nil
	default:
		return nil, errors.Errorf("unsupported memStrategy, expect dynlevel|none")
	}
	return &mm, nil
}

func validateInterval(interval int) error {
	if interval > 0 && interval <= constant.DefaultMaxMemCheckInterval {
		return nil
	}
	return errors.Errorf("check interval should between 0 and %v", constant.DefaultMemCheckInterval)
}

// Run wait every interval and execute run
func (m *MemoryManager) Run() {
	m.md.Run()
}

func writeMemoryLimit(cgroupPath string, limit string, ft fileType) error {
	var filename string
	switch ft {
	case mlimit:
		filename = memoryLimitFile
	case msoftLimit:
		filename = memorySoftLimitFile
	default:
		return errors.Errorf("unsupported file type %v", ft)
	}

	if err := writeMemoryFile(cgroupPath, filename, limit); err != nil {
		return errors.Errorf("limit memory failed for %s/%s=%s: %v", cgroupPath, filename, limit, err)
	}

	return nil
}

func writeMemoryFile(cgroupPath, filename, value string) error {
	cgFilePath, err := securejoin.SecureJoin(cgroupPath, filename)
	if err != nil {
		return errors.Errorf("join path failed for %s and %s: %v", cgroupPath, filename, err)
	}

	return ioutil.WriteFile(cgFilePath, []byte(value), constant.DefaultFileMode)
}

func readMemoryFile(path string) (int64, error) {
	const (
		base, width = 10, 64
	)
	content, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return 0, err
	}

	memBytes := strings.Split(string(content), "\n")[0]
	return strconv.ParseInt(memBytes, base, width)
}

// getMemoryInfo returns memory info
func getMemoryInfo() (memoryInfo, error) {
	var m memoryInfo
	var total, free, available int64
	const memInfoFile = "/proc/meminfo"

	f, err := os.Open(memInfoFile)
	if err != nil {
		return m, err
	}

	defer f.Close()

	// MemTotal:       15896176 kB
	// MemFree:         3811032 kB
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		if bytes.HasPrefix(scan.Bytes(), []byte("MemTotal:")) {
			if _, err := fmt.Sscanf(scan.Text(), "MemTotal:%d", &total); err != nil {
				return m, err
			}
		}

		if bytes.HasPrefix(scan.Bytes(), []byte("MemFree:")) {
			if _, err := fmt.Sscanf(scan.Text(), "MemFree:%d", &free); err != nil {
				return m, err
			}
		}

		if bytes.HasPrefix(scan.Bytes(), []byte("MemAvailable:")) {
			if _, err := fmt.Sscanf(scan.Text(), "MemAvailable:%d", &available); err != nil {
				return m, err
			}
		}
	}

	if total == 0 || free == 0 || available == 0 {
		return m, errors.Errorf("Memory value should be larger than 0, MemTotal:%d, MemFree:%d, MemAvailable:%d", total, free, available)
	}

	m.free = free * 1024
	m.total = total * 1024
	m.available = available * 1024

	return m, nil
}
