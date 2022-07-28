// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2021-04-17
// Description: This file contains default constants used in the project

// Package constant is for constant definition
package constant

import (
	"errors"
	"os"
	"time"
)

const (
	// RubikSock is path for rubik socket file
	RubikSock = "/run/rubik/rubik.sock"
	// ConfigFile is rubik config file
	ConfigFile = "/var/lib/rubik/config.json"
	// DefaultLogDir is default log dir
	DefaultLogDir = "/var/log/rubik"
	// LockFile is rubik lock file
	LockFile = "/run/rubik/rubik.lock"
	// ReadTimeout is timeout for http read
	ReadTimeout = 60 * time.Second
	// WriteTimeout is timeout for http write
	WriteTimeout = 60 * time.Second
	// DefaultSucceedCode is succeed code
	DefaultSucceedCode = 0
	// DefaultCgroupRoot is mount point
	DefaultCgroupRoot = "/sys/fs/cgroup"
	// CPUCgroupFileName is name of cgroup file used for cpu qos level setting
	CPUCgroupFileName = "cpu.qos_level"
	// MemoryCgroupFileName is name of cgroup file used for memory qos level setting
	MemoryCgroupFileName = "memory.qos_level"
	// DefaultFileMode is file mode for cgroup files
	DefaultFileMode os.FileMode = 0600
	// DefaultDirMode is dir default mode
	DefaultDirMode os.FileMode = 0700
	// DefaultUmask is default umask
	DefaultUmask = 0077
	// MaxCgroupPathLen is max cgroup path length for pod
	MaxCgroupPathLen = 4096
	// MaxPodIDLen is max pod id length
	MaxPodIDLen = 256
	// MaxPodsPerRequest is max pods number per http request
	MaxPodsPerRequest = 100
	// TmpTestDir is tmp directory for test
	TmpTestDir = "/tmp/rubik-test"
	// TaskChanCapacity is capacity for task chan
	TaskChanCapacity = 1024
	// WorkerNum is number of workers
	WorkerNum = 1
	// KubepodsCgroup is kubepods root cgroup
	KubepodsCgroup = "kubepods"
	// PodCgroupNamePrefix is pod cgroup name prefix
	PodCgroupNamePrefix = "pod"
	// NodeNameEnvKey is node name environment variable key
	NodeNameEnvKey = "RUBIK_NODE_NAME"
	// PriorityAnnotationKey is annotation key to mark offline pod
	PriorityAnnotationKey = "volcano.sh/preemptable"
	// CacheLimitAnnotationKey is annotation key to set L3/Mb resctrl group
	CacheLimitAnnotationKey = "volcano.sh/cache-limit"
	// QuotaBurstAnnotationKey is annotation key to set cpu.cfs_burst_ns
	QuotaBurstAnnotationKey = "volcano.sh/quota-burst-time"
	// BlkioKey is annotation key to set blkio limit
	BlkioKey = "volcano.sh/blkio-limit"
	// DefaultMemCheckInterval indicates the default memory check interval 5s.
	DefaultMemCheckInterval = 5
	// DefaultMaxMemCheckInterval indicates the default max memory check interval 30s.
	DefaultMaxMemCheckInterval = 30
	// DefaultMemStrategy indicates the default memory strategy.
	DefaultMemStrategy = "none"
)

// LevelType is type definition of qos level
type LevelType int32

const (
	// MinLevel is min level for qos level
	MinLevel LevelType = -1
	// MaxLevel is max level for qos level
	MaxLevel LevelType = 0
)

// Int is type casting for type LevelType
func (l LevelType) Int() int {
	return int(l)
}

const (
	// ErrCodeFailed for normal failed
	ErrCodeFailed = 1
)

// error define ref from src/internal/oserror/errors.go
var (
	// ErrFileTooBig file too big
	ErrFileTooBig = errors.New("file too big")
)
