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
	"path/filepath"
	"strings"

	"isula.org/rubik/pkg/common/constant"
	corev1 "k8s.io/api/core/v1"
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
		corev1.ContainerStatus
	}
	// RawPod represents kubernetes pod structure
	RawPod corev1.Pod
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

	switch pod.Status.QOSClass {
	case corev1.PodQOSGuaranteed:
		return filepath.Join(constant.KubepodsCgroup, constant.PodCgroupNamePrefix+id)
	case corev1.PodQOSBurstable:
		return filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)),
			constant.PodCgroupNamePrefix+id)
	case corev1.PodQOSBestEffort:
		return filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)),
			constant.PodCgroupNamePrefix+id)
	default:
		return ""
	}
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
			containerStatus,
		}
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
		id := rawContainer.getRealContainerID()
		if id == "" {
			continue
		}
		idContainersMap[id] = NewContainerInfo(rawContainer.Name, id, podCgroupPath)
	}
	return idContainersMap
}
