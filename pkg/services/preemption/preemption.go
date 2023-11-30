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
// Create: 2023-02-10
// Description: This file implement qos level setting service

// Package preemption is the service used for qos level setting
package preemption

import (
	"fmt"
	"strconv"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services/helper"
)

var supportCgroupTypes = map[string]*cgroup.Key{
	"cpu":    {SubSys: "cpu", FileName: constant.CPUCgroupFileName},
	"memory": {SubSys: "memory", FileName: constant.MemoryCgroupFileName},
}

// Preemption define service which related to qos level setting
type Preemption struct {
	helper.ServiceBase
	config PreemptionConfig
}

// PreemptionConfig define which resources need to use the preemption
type PreemptionConfig struct {
	Resource []string `json:"resource,omitempty"`
}

// PreemptionFactory is the factory os Preemption.
type PreemptionFactory struct {
	ObjName string
}

// Name to get the Preemption factory name.
func (i PreemptionFactory) Name() string {
	return "PreemptionFactory"
}

// NewObj to create object of Preemption.
func (i PreemptionFactory) NewObj() (interface{}, error) {
	return &Preemption{ServiceBase: helper.ServiceBase{Name: i.ObjName}}, nil
}

// SetConfig to config Preemption configure
func (q *Preemption) SetConfig(f helper.ConfigHandler) error {
	if f == nil {
		return fmt.Errorf("no config handler function callback")
	}

	var c PreemptionConfig
	if err := f(q.Name, &c); err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return err
	}
	q.config = c
	return nil
}

// PreStart is the pre-start action
func (q *Preemption) PreStart(viewer api.Viewer) error {
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	for _, pod := range viewer.ListPodsWithOptions() {
		if err := q.SetQoSLevel(pod); err != nil {
			log.Errorf("failed to set the qos level for the previously started pod %v: %v", pod.Name, err)
		}
	}
	return nil
}

// AddPod implement add function when pod is added in k8s
func (q *Preemption) AddPod(pod *typedef.PodInfo) error {
	if err := q.SetQoSLevel(pod); err != nil {
		return err
	}
	return q.validateConfig(pod)
}

// UpdatePod implement update function when pod info is changed
func (q *Preemption) UpdatePod(old, new *typedef.PodInfo) error {
	oldQos, newQos := getQoSLevel(old), getQoSLevel(new)
	switch {
	case newQos == oldQos:
		return nil
	case newQos > oldQos:
		return fmt.Errorf("does not support pod qos level setting from low to high")
	default:
		if err := q.validateConfig(new); err != nil {
			if err := q.SetQoSLevel(new); err != nil {
				return fmt.Errorf("failed to update the qos level of pod %s(%s): %v", new.Name, new.UID, err)
			}
		}
	}
	return nil
}

// DeletePod implement delete function when pod is deleted by k8s
func (q *Preemption) DeletePod(_ *typedef.PodInfo) error {
	return nil
}

// validateConfig will validate pod's qos level between value from
// cgroup file and the one from pod info
func (q *Preemption) validateConfig(pod *typedef.PodInfo) error {
	targetLevel := getQoSLevel(pod)
	for _, r := range q.config.Resource {
		if err := pod.GetCgroupAttr(supportCgroupTypes[r]).Expect(targetLevel); err != nil {
			return fmt.Errorf("failed to validate the qos level configuration of pod %s: %v", pod.Name, err)
		}
		for _, container := range pod.IDContainersMap {
			if err := container.GetCgroupAttr(supportCgroupTypes[r]).Expect(targetLevel); err != nil {
				return fmt.Errorf("failed to validate the qos level configuration of container %s: %v", pod.Name, err)
			}
		}
	}
	return nil
}

// SetQoSLevel set pod and all containers' qos level within it
func (q *Preemption) SetQoSLevel(pod *typedef.PodInfo) error {
	if pod == nil {
		return fmt.Errorf("empty pod info")
	}
	qosLevel := getQoSLevel(pod)
	if qosLevel == constant.Online {
		log.Infof("pod %s(%s) has already been set to online(%d)", pod.Name, pod.UID, qosLevel)
		return nil
	}

	var errs error
	for _, r := range q.config.Resource {
		if err := pod.SetCgroupAttr(supportCgroupTypes[r], strconv.Itoa(qosLevel)); err != nil {
			log.Warnf("failed to set %s-qos-level for pod %s: %v", r, pod.Name, err)
			errs = util.AppendErr(errs, err)
		}
		for _, container := range pod.IDContainersMap {
			if err := container.SetCgroupAttr(supportCgroupTypes[r], strconv.Itoa(qosLevel)); err != nil {
				log.Warnf("failed to set %s-qos-level for container %s: %v", r, container.Name, err)
				errs = util.AppendErr(errs, err)
			}
		}
	}
	if errs != nil {
		return errs
	}
	log.Infof("pod %s(%s) is set to offline(%d) successfully", pod.Name, pod.UID, qosLevel)
	return nil
}

func getQoSLevel(pod *typedef.PodInfo) int {
	if pod == nil {
		return constant.Online
	}
	if pod.Offline() {
		return constant.Offline
	}

	return constant.Online
}

// Validate will validate the qos service config
func (conf *PreemptionConfig) Validate() error {
	if len(conf.Resource) == 0 {
		return fmt.Errorf("empty resource preemption configuration")
	}
	for _, r := range conf.Resource {
		if _, ok := supportCgroupTypes[r]; !ok {
			return fmt.Errorf("does not support setting the %s subsystem", r)
		}
	}
	return nil
}
