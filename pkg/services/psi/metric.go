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
// Date: 2023-05-16
// Description: This file defines metrics used for psi service

package psi

import (
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/metric"
	"isula.org/rubik/pkg/core/trigger"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	cpuRes                = "cpu"
	memoryRes             = "memory"
	ioRes                 = "io"
	psiSubSys             = "cpuacct"
	defaultAvg10Threshold = 5.0
)

// supportResources is the supported resource type
var supportResources map[string]*cgroup.Key = map[string]*cgroup.Key{
	cpuRes:    {SubSys: psiSubSys, FileName: constant.PSICPUCgroupFileName},
	memoryRes: {SubSys: psiSubSys, FileName: constant.PSIMemoryCgroupFileName},
	ioRes:     {SubSys: psiSubSys, FileName: constant.PSIIOCgroupFileName},
}

// BasePSIMetric is the basic PSI indicator
type BasePSIMetric struct {
	*metric.BaseMetric
	avg10Threshold float64
	resources      []string
	// conservation is the Pod object that needs to guarantee resources
	conservation map[string]*typedef.PodInfo
	// Suspicion is the pod object that needs to be suspected of eviction
	suspicion map[string]*typedef.PodInfo
}

// Update updates the PSI CPU indicator of the cgroup list
func (m *BasePSIMetric) Update() error {
	if len(m.conservation) == 0 || len(m.suspicion) == 0 {
		log.Debugf("lack of guarantors or suspicious objects")
		return nil
	}
	for _, typ := range m.resources {
		if detectPSiMetric(typ, m.conservation, m.avg10Threshold) {
			if err := alarm(typ, m.Triggers, m.suspicion); err != nil {
				return err
			}
		}
	}
	return nil
}

func detectPSiMetric(resTyp string, conservation map[string]*typedef.PodInfo, avg10Threshold float64) bool {
	var key *cgroup.Key
	key, supported := supportResources[resTyp]
	if !supported {
		log.Errorf("undefined resource type %v", resTyp)
		return false
	}

	for _, pod := range conservation {
		log.Debugf("check psi of online pod: %v", pod.Name)
		pressure, err := pod.GetCgroupAttr(key).PSI()
		if err != nil {
			log.Warnf("fail to get file %v: %v", key.FileName, err)
			continue
		}
		if pressure.Some.Avg10 > avg10Threshold {
			log.Warnf("%v resource of pod %v reaches psi avg10 threshold (cur: %v, threshold: %v)",
				resTyp, pod.UID, pressure.Some.Avg10, avg10Threshold)
			return true
		}
	}
	return false
}

func alarm(resTyp string, triggers []trigger.Trigger, suspicion map[string]*typedef.PodInfo) error {
	var errs error
	const prefix = "max_"
	for _, t := range triggers {
		log.Infof("trigger %v", t.Name())
		if err := t.Execute(&trigger.FactorImpl{Msg: prefix + resTyp, Pods: suspicion}); err != nil {
			errs = util.AppendErr(errs, err)
		}
	}
	return errs
}
