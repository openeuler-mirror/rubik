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
	"os"
)

// the files and directories used by the system by default
const (
	// ConfigFile is rubik config file
	ConfigFile = "/var/lib/rubik/config.json"
	// LockFile is rubik lock file
	LockFile = "/run/rubik/rubik.lock"
	// DefaultCgroupRoot is mount point
	DefaultCgroupRoot = "/sys/fs/cgroup"
	// TmpTestDir is tmp directory for test
	TmpTestDir = "/tmp/rubik-test"
)

// kubernetes related configuration
const (
	// KubepodsCgroup is kubepods root cgroup
	KubepodsCgroup = "kubepods"
	// PodCgroupNamePrefix is pod cgroup name prefix
	PodCgroupNamePrefix = "pod"
	// NodeNameEnvKey is node name environment variable key
	NodeNameEnvKey = "RUBIK_NODE_NAME"
)

// File permission
const (
	// DefaultUmask is default umask
	DefaultUmask = 0077
	// DefaultFileMode is file mode for cgroup files
	DefaultFileMode os.FileMode = 0600
	// DefaultDirMode is dir default mode
	DefaultDirMode os.FileMode = 0700
	// DefaultDumpLogFileMode is the permission of the log file (recorded or archived)
	DefaultDumpLogFileMode os.FileMode = 0400
)

// Pod Annotation
const (
	// PriorityAnnotationKey is annotation key to mark offline pod
	PriorityAnnotationKey = "volcano.sh/preemptable"
	// CacheLimitAnnotationKey is annotation key to set L3/Mb resctrl group
	CacheLimitAnnotationKey = "volcano.sh/cache-limit"
	// QuotaBurstAnnotationKey is annotation key to set cpu.cfs_burst_ns
	QuotaBurstAnnotationKey = "volcano.sh/quota-burst-time"
	// BlkioKey is annotation key to set blkio limit
	BlkioKey = "volcano.sh/blkio-limit"
	// QuotaAnnotationKey is annotation key to mark whether to enable the quota turbo
	QuotaAnnotationKey = "volcano.sh/quota-turbo"
	// CpiAnnotationKey is annotation key to mark whether to enable the cpi
	CpiAnnotationKey = "volcano.sh/cpi"
)

// log config
const (
	LogDriverStdio  = "stdio"
	LogDriverFile   = "file"
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelWarn    = "warn"
	LogLevelError   = "error"
	LogLevelStack   = "stack"
	DefaultLogDir   = "/var/log/rubik"
	DefaultLogLevel = LogLevelInfo
	DefaultLogSize  = 1024
	// LogEntryKey is the key representing EntryName in the context
	LogEntryKey = "module"
)

// exit code
const (
	// NORMALEXIT for the normal exit code
	NormalExitCode int = iota
	// ArgumentErrorExitCode for normal failed
	ArgumentErrorExitCode
	// RepeatRunExitCode for repeat run exit
	RepeatRunExitCode
	// ErrorExitCode failed during run
	ErrorExitCode
)

// qos level
const (
	Offline = -1
	Online  = 0
)

// cgroup file name
const (
	// CPUCgroupFileName is name of cgroup file used for cpu qos level setting
	CPUCgroupFileName = "cpu.qos_level"
	// MemoryCgroupFileName is name of cgroup file used for memory qos level setting
	MemoryCgroupFileName = "memory.qos_level"
	// PSICPUCgroupFileName is name of cgroup file used for detecting cpu psi
	PSICPUCgroupFileName = "cpu.pressure"
	// PSIMemoryCgroupFileName is name of cgroup file used for detecting memory psi
	PSIMemoryCgroupFileName = "memory.pressure"
	// PSIIOCgroupFileName is name of cgroup file used for detecting io psi
	PSIIOCgroupFileName = "io.pressure"
)

// cgroup driver
const (
	// CgroupDriverSystemd is global config for cgroupfs driver choice:  systemd driver
	CgroupDriverSystemd = "systemd"
	// CgroupDriverCgroupfs is global config for cgroupfs driver choice:  cgroupfs driver
	CgroupDriverCgroupfs = "cgroupfs"
)

// container engine
const (
	// ContainerEngineCrio is name of crio container engine
	ContainerEngineCrio = "crio"
	// ContainerEngineContainerd is name of containerd container engine
	ContainerEngineContainerd = "cri-containerd"
	// ContainerEngineIsula is name of isula container engine
	ContainerEngineIsula = "isula"
	// ContainerEngineDocker is name of docker container engine
	ContainerEngineDocker = "docker"
)

// informer type
const (
	// APIServerInformer is global config for informerType choice: informerType: "apiserver"
	APIServerInformer = "apiserver"
	// NRIInformer is global config for informerType choice: informerType: "nri"
	NRIInformer = "nri"
)
