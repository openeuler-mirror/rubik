// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2023-01-05
// Description: This file defines podInfo

// Package typedef defines core struct and methods for rubik
package typedef

import (
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// PodInfo represents pod
type PodInfo struct {
	cgroup.Hierarchy
	Name                string                    `json:"name"`
	UID                 string                    `json:"uid"`
	Namespace           string                    `json:"namespace"`
	IDContainersMap     map[string]*ContainerInfo `json:"containers,omitempty"`
	Annotations         map[string]string         `json:"annotations,omitempty"`
	Labels              map[string]string         `json:"labels,omitempty"`
	ID                  string                    `json:"id,omitempty"` // id of the sandbox container in pod
	nriContainerRequest map[string]ResourceMap
	nriContainerLimit   map[string]ResourceMap
}

// NewPodInfo creates the PodInfo instance
func NewPodInfo(pod *RawPod) *PodInfo {
	return &PodInfo{
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		UID:             pod.ID(),
		Hierarchy:       cgroup.Hierarchy{Path: pod.CgroupPath()},
		IDContainersMap: pod.ExtractContainerInfos(),
		Annotations:     pod.DeepCopy().Annotations,
		Labels:          pod.DeepCopy().Labels,
	}
}

// DeepCopy returns deepcopy object
func (pod *PodInfo) DeepCopy() *PodInfo {
	if pod == nil {
		return nil
	}

	var copy = *pod
	// nil is different from empty value in golang
	if pod.IDContainersMap != nil {
		contMap := make(map[string]*ContainerInfo)
		for id, cont := range pod.IDContainersMap {
			contMap[id] = cont.DeepCopy()
		}
		copy.IDContainersMap = contMap
	}

	if pod.Annotations != nil {
		annoMap := make(map[string]string)
		for k, v := range pod.Annotations {
			annoMap[k] = v
		}
		copy.Annotations = annoMap
	}

	if pod.Labels != nil {
		labelMap := make(map[string]string)
		for k, v := range pod.Labels {
			labelMap[k] = v
		}
		copy.Labels = labelMap
	}

	if pod.nriContainerLimit != nil {
		limits := make(map[string]ResourceMap)
		for k, v := range pod.nriContainerLimit {
			limits[k] = v.DeepCopy()
		}
		copy.nriContainerLimit = limits
	}

	if pod.nriContainerRequest != nil {
		requests := make(map[string]ResourceMap)
		for k, v := range pod.nriContainerRequest {
			requests[k] = v.DeepCopy()
		}
		copy.nriContainerRequest = requests
	}
	return &copy
}

// Offline is used to determine whether the pod is offline
func (pod *PodInfo) Offline() bool {
	var anno string
	var label string

	if pod.Annotations != nil {
		anno = pod.Annotations[constant.PriorityAnnotationKey]
	}

	if pod.Labels != nil {
		label = pod.Labels[constant.PriorityAnnotationKey]
	}

	// Annotations have a higher priority than labels
	return anno == "true" || label == "true"
}

// Online is used to determine whether the pod is online
func (pod *PodInfo) Online() bool {
	return !pod.Offline()
}

func (pod *PodInfo) GetNriContainerRequest() map[string]ResourceMap {
	return pod.nriContainerRequest
}

func (pod *PodInfo) GetNriContainerLimit() map[string]ResourceMap {
	return pod.nriContainerLimit
}
