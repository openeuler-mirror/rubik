// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-10-31
// Description: This file is used for resource Analyzer

package analyze

import (
	"runtime"

	v2 "github.com/google/cadvisor/info/v2"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/resource/manager/common"
)

type Calculator func(*typedef.PodInfo) float64

type Analyzer struct {
	common.Manager
}

func NewResourceAnalyzer(manager common.Manager) *Analyzer {
	return &Analyzer{
		Manager: manager,
	}
}

func (a *Analyzer) CPUCalculatorBuilder(reqOpt *common.GetOption) Calculator {
	return func(pi *typedef.PodInfo) float64 {
		const (
			miniNum        int     = 2
			nanoToMicro    float64 = 1000
			percentageRate float64 = 100
		)

		podStats := a.getPodStats("/"+pi.Path, reqOpt)
		if len(podStats) < miniNum {
			log.Errorf("pod %v has no enough cpu stats collected, skip it", pi.Name)
			return -1
		}
		var (
			last        = podStats[len(podStats)-1]
			penultimate = podStats[len(podStats)-2]
			cpuUsageUs  = float64(last.Cpu.Usage.Total-penultimate.Cpu.Usage.Total) / nanoToMicro
			timeDeltaUs = float64(last.Timestamp.Sub(penultimate.Timestamp).Microseconds())
		)
		return util.Div(cpuUsageUs, timeDeltaUs) / float64(runtime.NumCPU()) * percentageRate
	}
}

func (a *Analyzer) MemoryCalculatorBuilder(reqOpt *common.GetOption) Calculator {
	return func(pi *typedef.PodInfo) float64 {
		const (
			bytesToMb float64 = 1000000.0
			miniNum   int     = 1
		)
		podStats := a.getPodStats("/"+pi.Path, reqOpt)
		if len(podStats) < miniNum {
			return -1
		}
		return float64(podStats[len(podStats)-1].Memory.Usage) / bytesToMb
	}
}

func (a *Analyzer) getPodStats(cgroupPath string, reqOpt *common.GetOption) []*v2.ContainerStats {
	infoMap, err := a.GetPodStats(cgroupPath, *reqOpt)
	if err != nil {
		log.Warnf("failed to get cgroup information %v: %v", cgroupPath, err)
		return nil
	}
	info, existed := infoMap[cgroupPath]
	if !existed {
		log.Warnf("failed to get cgroup %v from cadvisor", cgroupPath)
		return nil
	}
	return info.Stats
}
