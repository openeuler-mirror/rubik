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

// Package qos is the service used for qos level setting
package preemption

import (
	"fmt"
	"strconv"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services"
)

var supportCgroupTypes = map[string]*cgroup.Key{
	"cpu":    {SubSys: "cpu", FileName: constant.CPUCgroupFileName},
	"memory": {SubSys: "memory", FileName: constant.MemoryCgroupFileName},
}

// Preemption define service which related to qos level setting
type Preemption struct {
	Name string `json:"-"`
	Config
}

// Config contains sub-system that need to set qos level
type Config struct {
	Resource []string `json:"resource,omitempty"`
}

func init() {
	services.Register("qos", func() interface{} {
		return NewQoS()
	})
}

// NewQoS return qos instance
func NewQoS() *Preemption {
	return &Preemption{
		Name: "qos",
	}
}

// ID return qos service name
func (q *Preemption) ID() string {
	return q.Name
}

// PreStart is the pre-start action
func (q *Preemption) PreStart(viewer api.Viewer) error {
	for _, pod := range viewer.ListPodsWithOptions() {
		if err := q.SetQoSLevel(pod); err != nil {
			log.Errorf("error prestart pod %v: %v", pod.Name, err)
		}
	}
	return nil
}

// AddFunc implement add function when pod is added in k8s
func (q *Preemption) AddFunc(pod *typedef.PodInfo) error {
	if err := q.SetQoSLevel(pod); err != nil {
		return err
	}
	if err := q.ValidateConfig(pod); err != nil {
		return err
	}
	return nil
}

// UpdateFunc implement update function when pod info is changed
func (q *Preemption) UpdateFunc(old, new *typedef.PodInfo) error {
	oldQos, newQos := getQoSLevel(old), getQoSLevel(new)
	switch {
	case newQos == oldQos:
		return nil
	case newQos > oldQos:
		return fmt.Errorf("not support change qos level from low to high")
	default:
		if err := q.ValidateConfig(new); err != nil {
			if err := q.SetQoSLevel(new); err != nil {
				return fmt.Errorf("update qos for pod %s(%s) failed: %v", new.Name, new.UID, err)
			}
		}
	}
	return nil
}

// DeleteFunc implement delete function when pod is deleted by k8s
func (q *Preemption) DeleteFunc(pod *typedef.PodInfo) error {
	return nil
}

// ValidateConfig will validate pod's qos level between value from
// cgroup file and the one from pod info
func (q *Preemption) ValidateConfig(pod *typedef.PodInfo) error {
	targetLevel := getQoSLevel(pod)
	for _, r := range q.Resource {
		if err := pod.GetCgroupAttr(supportCgroupTypes[r]).Expect(targetLevel); err != nil {
			return fmt.Errorf("failed to validate pod %s: %v", pod.Name, err)
		}
		for _, container := range pod.IDContainersMap {
			if err := container.GetCgroupAttr(supportCgroupTypes[r]).Expect(targetLevel); err != nil {
				return fmt.Errorf("failed to validate pod %s: %v", pod.Name, err)
			}
		}
	}
	return nil
}

// SetQoSLevel set pod and all containers' qos level within it
func (q *Preemption) SetQoSLevel(pod *typedef.PodInfo) error {
	if pod == nil {
		return fmt.Errorf("pod info is empty")
	}
	qosLevel := getQoSLevel(pod)
	if qosLevel == constant.Online {
		log.Debugf("pod %s already online", pod.Name)
		return nil
	}

	for _, r := range q.Resource {
		if err := pod.SetCgroupAttr(supportCgroupTypes[r], strconv.Itoa(qosLevel)); err != nil {
			return err
		}
		for _, container := range pod.IDContainersMap {
			if err := container.SetCgroupAttr(supportCgroupTypes[r], strconv.Itoa(qosLevel)); err != nil {
				return err
			}
		}
	}
	log.Debugf("set pod %s(%s) qos level %d ok", pod.Name, pod.UID, qosLevel)
	return nil
}

func getQoSLevel(pod *typedef.PodInfo) int {
	if pod == nil {
		return constant.Online
	}
	anno, ok := pod.Annotations[constant.PriorityAnnotationKey]
	if !ok {
		return constant.Online
	}
	switch anno {
	case "true":
		return constant.Offline
	case "false":
		return constant.Online
	default:
		return constant.Online
	}
}

// Validate will validate the qos service config
func (q *Preemption) Validate() error {
	if len(q.Resource) == 0 {
		return fmt.Errorf("empty qos config")
	}
	for _, r := range q.Resource {
		if _, ok := supportCgroupTypes[r]; !ok {
			return fmt.Errorf("not support sub system %s", r)
		}
	}
	return nil
}