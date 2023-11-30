// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-05-18
// Description: This file is used for resource Analyzer

package trigger

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/cadvisor/cache/memory"
	"github.com/google/cadvisor/container"
	v2 "github.com/google/cadvisor/info/v2"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/utils/sysfs"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/resourcemanager/cadvisor"
)

const (
	miniNum                = 2
	nanoToMicro    float64 = 1000
	percentageRate float64 = 100
)

// resourceAnalysisExec is the singleton of Analyzer executor implementation
var resourceAnalysisExec *Analyzer

// analyzerCreator creates Analyzer trigger
var analyzerCreator = func() Trigger {
	if resourceAnalysisExec == nil {
		m := NewCadvisorManager()
		if m != nil {
			log.Infof("initialize resourceAnalysisExec")
			resourceAnalysisExec = &Analyzer{cadvisorManager: m}
			appendUsedExecutors(ResourceAnalysisAnno, resourceAnalysisExec)
		}
	}
	return withTreeTrigger(ResourceAnalysisAnno, resourceAnalysisExec)
}

// rreqOpt is the option to get information from cadvisor
var reqOpt = v2.RequestOptions{
	IdType:    v2.TypeName,
	Count:     2,
	Recursive: false,
}

// Analyzer is the resource analysis trigger
type Analyzer struct {
	sync.RWMutex
	cadvisorManager *cadvisor.Manager
}

// NewCadvisorManager returns an cadvisor.Manager object
func NewCadvisorManager() *cadvisor.Manager {
	const (
		cacheMinutes       = 10
		keepingIntervalSec = 10
	)
	var (
		allowDynamic            = true
		maxHousekeepingInterval = time.Duration(keepingIntervalSec * time.Second)
		cacheAge                = time.Duration(cacheMinutes * time.Minute)
	)
	args := cadvisor.StartArgs{
		MemCache: memory.New(cacheAge, nil),
		SysFs:    sysfs.NewRealSysFs(),
		IncludeMetrics: container.MetricSet{
			container.CpuUsageMetrics:    struct{}{},
			container.MemoryUsageMetrics: struct{}{},
			container.DiskUsageMetrics:   struct{}{},
			container.DiskIOMetrics:      struct{}{},
		},
		MaxHousekeepingConfig: manager.HouskeepingConfig{
			Interval:     &maxHousekeepingInterval,
			AllowDynamic: &allowDynamic,
		},
	}
	return cadvisor.WithStartArgs(args)
}

// Execute filters the corresponding Pod according to the operation type and triggers it on demand
func (a *Analyzer) Execute(f Factor) (Factor, error) {
	a.RLock()
	defer a.RUnlock()
	if a.cadvisorManager == nil {
		return nil, fmt.Errorf("failed to use cadvisor, please check")
	}
	var (
		target *typedef.PodInfo
		opTyp  = f.Message()
		errMsg string
		alarm  = func(target *typedef.PodInfo, errMsg string) (Factor, error) {
			if target == nil {
				return nil, fmt.Errorf(errMsg)
			}
			return &FactorImpl{Pods: map[string]*typedef.PodInfo{target.Name: target}}, nil
		}
	)
	log.Debugf("receive operation: %v", opTyp)
	switch opTyp {
	case "max_cpu":
		errMsg = "unable to find pod with maximum CPU utilization"
		target = a.maxCPUUtil(f.TargetPods())
	case "max_memory":
		errMsg = "unable to find pod with maximum memory utilization"
		target = a.maxMemoryUtil(f.TargetPods())
	case "max_io":
		errMsg = "unable to find pod with maximum I/O bandwidth"
		target = a.maxCPUUtil(f.TargetPods())
	default:
		errMsg = "undefined operation: " + opTyp
	}
	return alarm(target, errMsg)
}

func (a *Analyzer) maxCPUUtil(pods map[string]*typedef.PodInfo) *typedef.PodInfo {
	var (
		chosen  *typedef.PodInfo
		maxUtil float64
	)
	for name, pod := range pods {
		podStats, err := a.cgroupCadvisorInfo("/"+pod.Path, reqOpt)
		if err != nil {
			log.Errorf("failed to get cgroup information %v: %v", pod.Path, err)
			continue
		}
		if len(podStats) < miniNum {
			log.Errorf("pod %v has no enough cpu stats collected, skip it", name)
			continue
		}
		last := podStats[len(podStats)-1]
		penultimate := podStats[len(podStats)-2]
		cpuUsageUs := float64(last.Cpu.Usage.Total-penultimate.Cpu.Usage.Total) / nanoToMicro
		timeDeltaUs := float64(last.Timestamp.Sub(penultimate.Timestamp).Microseconds())
		cpuUtil := util.Div(cpuUsageUs, timeDeltaUs) * percentageRate
		log.Debugf("pod %v cpu util %v%%=%v/%v(us)", name, cpuUtil, cpuUsageUs, timeDeltaUs)
		if maxUtil < cpuUtil {
			maxUtil = cpuUtil
			chosen = pod
		}
	}
	if chosen != nil {
		log.Infof("find the pod(%v) with the highest cpu utilization(%v)", chosen.Name, maxUtil)
	}
	return chosen
}

func (a *Analyzer) maxMemoryUtil(pods map[string]*typedef.PodInfo) *typedef.PodInfo {
	var (
		chosen  *typedef.PodInfo
		maxUtil uint64
	)
	for name, pod := range pods {
		podStats, err := a.cgroupCadvisorInfo("/"+pod.Path, reqOpt)
		if err != nil {
			log.Errorf("failed to get cgroup information %v: %v", pod.Path, err)
			continue
		}
		last := podStats[len(podStats)-1].Memory.Usage
		log.Debugf("pod %v memory usage %vB", name, last)
		if maxUtil < last {
			maxUtil = last
			chosen = pod
		}
	}
	if chosen != nil {
		log.Infof("find the pod(%v) with the highest memory utilization(%v)", chosen.Name, maxUtil)
	}
	return chosen
}

func (a *Analyzer) maxIOBandwidth(_ map[string]*typedef.PodInfo) *typedef.PodInfo {
	return nil
}

func (a *Analyzer) cgroupCadvisorInfo(cgroupPath string, opts v2.RequestOptions) ([]*v2.ContainerStats, error) {
	infoMap, err := a.cadvisorManager.ContainerInfoV2(cgroupPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get cgroup information %v: %v", cgroupPath, err)
	}
	info, existed := infoMap[cgroupPath]
	if !existed {
		return nil, fmt.Errorf("failed to get cgroup info from cadvisor")
	}
	return info.Stats, nil
}

// Stop stops Analyzer
func (a *Analyzer) Stop() error {
	a.Lock()
	defer a.Unlock()
	m := a.cadvisorManager
	a.cadvisorManager = nil
	return m.Stop()
}
