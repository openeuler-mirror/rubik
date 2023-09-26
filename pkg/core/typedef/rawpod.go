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
// Description: This file defines RawPod which encapsulate kubernetes pods

// Package typedef defines core struct and methods for rubik
package typedef

import (
	"fmt"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"isula.org/rubik/pkg/common/constant"
)

const (
	configHashAnnotationKey = "kubernetes.io/config.hash"
	// RUNNING means the Pod is in the running phase
	RUNNING = corev1.PodRunning
)

type (
	// RawContainer is kubernetes contaienr structure
	RawContainer struct {
		/*
			The container information of kubernetes will be stored in pod.Status.ContainerStatuses
			and pod.Spec.Containers respectively.
			The container ID information is stored in ContainerStatuses.
			Currently our use of container information is limited to ID and Name,
			so we only implemented a simple RawContainer structure.
			You can continue to expand RawContainer in the future,
			such as saving the running state of the container.
		*/
		status corev1.ContainerStatus
		spec   corev1.Container
	}
	// RawPod represents kubernetes pod structure
	RawPod corev1.Pod
	// ResourceType indicates the resource type, such as memory or CPU
	ResourceType uint8
	// ResourceMap represents the available value of a certain type of resource
	ResourceMap map[ResourceType]float64
)

const (
	// ResourceCPU indicates CPU resources
	ResourceCPU ResourceType = iota
	// ResourceMem represents memory resources
	ResourceMem
)

// ExtractPodInfo returns podInfo from RawPod
func (pod *RawPod) ExtractPodInfo() *PodInfo {
	if pod == nil {
		return nil
	}
	return NewPodInfo(pod)
}

// Running returns true when pod is in the running phase
func (pod *RawPod) Running() bool {
	if pod == nil {
		return false
	}
	return pod.Status.Phase == RUNNING
}

// ID returns the unique identity of pod
func (pod *RawPod) ID() string {
	if pod == nil {
		return ""
	}
	return string(pod.UID)
}

// CgroupPath returns cgroup path of raw pod
func (pod *RawPod) CgroupPath() string {
	id := string(pod.UID)
	if configHash := pod.Annotations[configHashAnnotationKey]; configHash != "" {
		id = configHash
	}

	qosClassPath := ""
	switch pod.Status.QOSClass {
	case corev1.PodQOSGuaranteed:
	case corev1.PodQOSBurstable:
		qosClassPath = strings.ToLower(string(corev1.PodQOSBurstable))
	case corev1.PodQOSBestEffort:
		qosClassPath = strings.ToLower(string(corev1.PodQOSBestEffort))
	default:
		return ""
	}
	/*
		example:
		1. Burstable: pod requests are less than the value of limits and not 0;
		kubepods/burstable/pod34152897-dbaf-11ea-8cb9-0653660051c3
		2. BestEffort: pod requests and limits are both 0;
		kubepods/bestEffort/pod34152897-dbaf-11ea-8cb9-0653660051c3
		3. Guaranteed: pod requests are equal to the value set by limits;
		kubepods/pod34152897-dbaf-11ea-8cb9-0653660051c3
	*/
	return filepath.Join(constant.KubepodsCgroup, qosClassPath, constant.PodCgroupNamePrefix+id)
}

// ListRawContainers returns all RawContainers in the RawPod
func (pod *RawPod) ListRawContainers() map[string]*RawContainer {
	if pod == nil {
		return nil
	}
	var nameRawContainersMap = make(map[string]*RawContainer)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Since corev1.Container only exists the container name, use Name as the unique key
		nameRawContainersMap[containerStatus.Name] = &RawContainer{
			status: containerStatus,
		}
	}
	for _, container := range pod.Spec.Containers {
		cont, ok := nameRawContainersMap[container.Name]
		if !ok {
			continue
		}
		cont.spec = container
	}
	return nameRawContainersMap
}

// ExtractContainerInfos returns container information from Pod
func (pod *RawPod) ExtractContainerInfos() map[string]*ContainerInfo {
	var idContainersMap = make(map[string]*ContainerInfo, 0)
	// 1. get list of raw containers
	nameRawContainersMap := pod.ListRawContainers()
	if len(nameRawContainersMap) == 0 {
		return idContainersMap
	}

	// 2. generate ID-Container mapping
	podCgroupPath := pod.CgroupPath()
	for _, rawContainer := range nameRawContainersMap {
		id, err := rawContainer.GetRealContainerID()
		if err != nil {
			fmt.Printf("failed to parse container ID: %v\n", err)
			continue
		}
		if id == "" {
			continue
		}
		idContainersMap[id] = NewContainerInfo(id, podCgroupPath, rawContainer)
	}
	return idContainersMap
}

// GetRealContainerID parses the containerID of k8s
func (cont *RawContainer) GetRealContainerID() (string, error) {
	/*
		Note:
		An UNDEFINED container engine was used when the function was executed for the first time
		it seems unlikely to support different container engines at runtime,
		So we don't consider the case of midway container engine changes
		`fixContainerEngine` is only executed when `getRealContainerID` is called for the first time
	*/
	setContainerEnginesOnce.Do(func() { fixContainerEngine(cont.status.ContainerID) })

	if !currentContainerEngines.Support(cont) {
		return "", fmt.Errorf("unsupported container engine: %v", cont.status.ContainerID)
	}

	cid := cont.status.ContainerID[len(currentContainerEngines.Prefix()):]
	// the container may be in the creation or deletion phase.
	if len(cid) == 0 {
		return "", nil
	}
	return cid, nil
}

// GetResourceMaps returns the number of requests and limits of CPU and memory resources
func (cont *RawContainer) GetResourceMaps() (ResourceMap, ResourceMap) {
	const milli float64 = 1000
	var (
		// high precision
		converter = func(value *resource.Quantity) float64 {
			return float64(value.MilliValue()) / milli
		}
		iterator = func(resourceItems *corev1.ResourceList) ResourceMap {
			results := make(ResourceMap)
			results[ResourceCPU] = converter(resourceItems.Cpu())
			results[ResourceMem] = converter(resourceItems.Memory())
			return results
		}
	)
	return iterator(&cont.spec.Resources.Requests), iterator(&cont.spec.Resources.Limits)
}

// DeepCopy returns the deep copy object of ResourceMap
func (m ResourceMap) DeepCopy() ResourceMap {
	if m == nil {
		return nil
	}
	res := make(ResourceMap, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
