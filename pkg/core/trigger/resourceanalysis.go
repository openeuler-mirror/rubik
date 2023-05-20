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
// Description: This file is used for *

package trigger

import (
	"fmt"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
)

// resourceAnalysisExec is the singleton of Analyzer triggers
var resourceAnalysisExec = &Analyzer{}
var analyzerCreator = func() Trigger {
	return &TreeTrigger{name: ResourceAnalysisAnno, exec: resourceAnalysisExec}
}

// Analyzer is the resource analysis trigger
type Analyzer struct{}

// Execute filters the corresponding Pod according to the operation type and triggers it on demand
func (a *Analyzer) Execute(f Factor) (Factor, error) {
	var (
		target string
		opTyp  = f.Message()
		errMsg string
	)
	log.Infof("receive operation: %v", opTyp)
	alarm := func(target, errMsg string) (Factor, error) {
		if len(target) == 0 {
			return nil, fmt.Errorf(errMsg)
		}
		return &FactorImpl{Msg: target}, nil
	}
	switch opTyp {
	case "max_cpu":
		errMsg = "unable to find pod with maximum CPU utilization"
		target = a.maxCPUUtil(f.TargetPods())
	case "max_memory":
		errMsg = "unable to find pod with maximum memory utilization"
		target = a.maxMemoryUtil(f.TargetPods())
	case "max_io":
		errMsg = "unable to find pod with maximum I/O bandwidth"
		target = a.maxIOBandwidth(f.TargetPods())
	default:
		errMsg = "undefined operation: " + opTyp
	}
	return alarm(target, errMsg)
}

func (a *Analyzer) maxCPUUtil(pods map[string]*typedef.PodInfo) string {
	return "testtest"
}

func (a *Analyzer) maxMemoryUtil(pods map[string]*typedef.PodInfo) string {
	return ""
}

func (a *Analyzer) maxIOBandwidth(pods map[string]*typedef.PodInfo) string {
	return ""
}
