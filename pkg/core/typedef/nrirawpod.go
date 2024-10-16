// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: weiyuan
// Create: 2024-05-28
// Description:  This file defines NRIRawPod and NRIRawContainer which encapsulate nri pod and container info

// Package typedef defines core struct and methods for rubik
package typedef

import (
	"encoding/json"
	"fmt"

	"github.com/containerd/nri/pkg/api"
	v1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/core/typedef/cgroup"
)

type (
	// NRIRawContainer is nri container structure
	NRIRawContainer api.Container
	// NRIRawPod is nri pod structure
	NRIRawPod api.PodSandbox
)

// convert NRIRawPod structure to PodInfo structure
func (pod *NRIRawPod) ConvertNRIRawPod2PodInfo() *PodInfo {
	if pod == nil {
		return nil
	}
	requests, limits := pod.GetResourceMaps()
	return &PodInfo{
		Hierarchy: cgroup.Hierarchy{
			Path: pod.Linux.CgroupParent,
		},
		Name:                pod.Name,
		UID:                 pod.Uid,
		Namespace:           pod.Namespace,
		IDContainersMap:     make(map[string]*ContainerInfo, 0),
		Annotations:         pod.Annotations,
		Labels:              pod.Labels,
		ID:                  pod.Id,
		nriContainerRequest: requests,
		nriContainerLimit:   limits,
	}
}

// get pod running state
func (pod *NRIRawPod) Running() bool {
	return true
}

// get pod UID
func (pod *NRIRawPod) ID() string {
	if pod == nil {
		return ""
	}
	return string(pod.Uid)
}

func (pod *NRIRawPod) GetResourceMaps() (map[string]ResourceMap, map[string]ResourceMap) {
	const containerAppliedConfiguration = "kubectl.kubernetes.io/last-applied-configuration"
	configurations := pod.Annotations[containerAppliedConfiguration]
	if configurations == "" {
		fmt.Printf("empty resource map in pod %v\n", pod.Uid)
		return nil, nil
	}

	rawPod := &RawPod{}
	err := json.Unmarshal([]byte(configurations), rawPod)
	if err != nil {
		fmt.Printf("failed to unmarshal resource map %v: %v\n", configurations, err)
		return nil, nil
	}
	var requests, limits = map[string]ResourceMap{}, map[string]ResourceMap{}

	resourceMapConvert := func(rl v1.ResourceList) ResourceMap {
		const milli = 1000
		return ResourceMap{
			ResourceCPU: float64(rl.Cpu().MilliValue()) / milli,
			ResourceMem: float64(rl.Memory().MilliValue()) / milli, // memory size in bytes
		}
	}

	for _, cont := range rawPod.Spec.Containers {
		requests[cont.Name] = resourceMapConvert(cont.Resources.Requests)
		limits[cont.Name] = resourceMapConvert(cont.Resources.Limits)
	}
	return requests, limits
}
