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
	"path/filepath"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/common/constant"
)

// PodInfo represents pod
type PodInfo struct {
	// Basic Information
	Containers map[string]*ContainerInfo `json:"containers,omitempty"`
	Name       string                    `json:"name"`
	UID        string                    `json:"uid"`
	CgroupPath string                    `json:"cgroupPath"`
	Namespace  string                    `json:"namespace"`
	// TODO: elimate cgroupRoot
	CgroupRoot string `json:"cgroupRoot"`

	// TODO: use map[string]interface to replace annotations
	// Service Information
	Offline         bool   `json:"offline"`
	CacheLimitLevel string `json:"cacheLimitLevel,omitempty"`

	// value of quota burst
	QuotaBurst int64 `json:"quotaBurst"`
}

// NewPodInfo creates the PodInfo instance
func NewPodInfo(pod *RawPod, cgroupRoot string) *PodInfo {
	pi := &PodInfo{
		Name:       pod.Name,
		UID:        string(pod.UID),
		Containers: make(map[string]*ContainerInfo, 0),
		CgroupPath: GetPodCgroupPath(pod),
		Namespace:  pod.Namespace,
		CgroupRoot: cgroupRoot,
	}
	updatePodInfoNoLock(pi, pod)
	return pi
}

// updatePodInfoNoLock updates PodInfo from the pod of Kubernetes.
// UpdatePodInfoNoLock does not lock pods during the modification.
// Therefore, ensure that the pod is being used only by this function.
// Currently, the checkpoint manager variable is locked when this function is invoked.
func updatePodInfoNoLock(pi *PodInfo, pod *RawPod) {
	const (
		dockerPrefix     = "docker://"
		containerdPrefix = "containerd://"
	)
	pi.Name = pod.Name
	pi.Offline = IsOffline(pod)
	pi.CacheLimitLevel = GetPodCacheLimit(pod)
	pi.QuotaBurst = GetQuotaBurst(pod)

	nameID := make(map[string]string, len(pod.Status.ContainerStatuses))
	for _, c := range pod.Status.ContainerStatuses {
		// rubik is compatible with dockerd and containerd container engines.
		cid := strings.TrimPrefix(c.ContainerID, dockerPrefix)
		cid = strings.TrimPrefix(cid, containerdPrefix)

		// the container may be in the creation or deletion phase.
		if len(cid) == 0 {
			// log.Debugf("no container id found of container %v", c.Name)
			continue
		}
		nameID[c.Name] = cid
	}
	// update ContainerInfo in a PodInfo
	for _, c := range pod.Spec.Containers {
		ci, ok := pi.Containers[c.Name]
		// add a container
		if !ok {
			// log.Debugf("add new container %v", c.Name)
			pi.AddContainerInfo(NewContainerInfo(&c, string(pod.UID), nameID[c.Name],
				pi.CgroupRoot, pi.CgroupPath))
			continue
		}
		// The container name remains unchanged, and other information about the container is updated.
		ci.ID = nameID[c.Name]
		ci.CgroupAddr = filepath.Join(pi.CgroupPath, ci.ID)
	}
	// delete a container that does not exist
	for name := range pi.Containers {
		if _, ok := nameID[name]; !ok {
			// log.Debugf("delete container %v", name)
			delete(pi.Containers, name)
		}
	}
}

// Clone returns deepcopy object
func (pi *PodInfo) Clone() *PodInfo {
	if pi == nil {
		return nil
	}
	copy := *pi
	// deepcopy reference object
	copy.Containers = make(map[string]*ContainerInfo, len(pi.Containers))
	for _, c := range pi.Containers {
		copy.Containers[c.Name] = c.Clone()
	}
	return &copy
}

// AddContainerInfo add container info to pod
func (pi *PodInfo) AddContainerInfo(containerInfo *ContainerInfo) {
	// key should not be empty
	if containerInfo.Name == "" {
		return
	}
	pi.Containers[containerInfo.Name] = containerInfo
}

const configHashAnnotationKey = "kubernetes.io/config.hash"

// IsOffline judges whether pod is offline pod
func IsOffline(pod *RawPod) bool {
	return pod.Annotations[constant.PriorityAnnotationKey] == "true"
}

// GetPodCacheLimit returns cachelimit annotation
func GetPodCacheLimit(pod *RawPod) string {
	return pod.Annotations[constant.CacheLimitAnnotationKey]
}

// ParseInt64 converts the string type to Int64
func ParseInt64(str string) (int64, error) {
	const (
		base    = 10
		bitSize = 64
	)
	return strconv.ParseInt(str, base, bitSize)
}

// GetQuotaBurst checks CPU quota burst annotation value.
func GetQuotaBurst(pod *RawPod) int64 {
	quota := pod.Annotations[constant.QuotaBurstAnnotationKey]
	if quota == "" {
		return constant.InvalidBurst
	}

	quotaBurst, err := ParseInt64(quota)
	if err != nil {
		return constant.InvalidBurst
	}
	if quotaBurst < 0 {
		return constant.InvalidBurst
	}
	return quotaBurst
}

// GetPodCgroupPath returns cgroup path of pod
func GetPodCgroupPath(pod *RawPod) string {
	var cgroupPath string
	id := string(pod.UID)
	if configHash := pod.Annotations[configHashAnnotationKey]; configHash != "" {
		id = configHash
	}

	switch pod.Status.QOSClass {
	case corev1.PodQOSGuaranteed:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, constant.PodCgroupNamePrefix+id)
	case corev1.PodQOSBurstable:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)),
			constant.PodCgroupNamePrefix+id)
	case corev1.PodQOSBestEffort:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)),
			constant.PodCgroupNamePrefix+id)
	}

	return cgroupPath
}
