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

// PodInfo represents pod
type PodInfo struct {
	IDContainersMap map[string]*ContainerInfo `json:"containers,omitempty"`
	Name            string                    `json:"name"`
	UID             string                    `json:"uid"`
	CgroupPath      string                    `json:"cgroupPath"`
	Namespace       string                    `json:"namespace"`
	Annotations     map[string]string         `json:"annotations,omitempty"`
}

// NewPodInfo creates the PodInfo instance
func NewPodInfo(pod *RawPod) *PodInfo {
	return &PodInfo{
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		UID:             pod.ID(),
		CgroupPath:      pod.CgroupPath(),
		IDContainersMap: pod.ExtractContainerInfos(),
		Annotations:     pod.DeepCopy().Annotations,
	}
}

// DeepCopy returns deepcopy object
func (pi *PodInfo) DeepCopy() *PodInfo {
	if pi == nil {
		return nil
	}
	// deepcopy reference object
	idContainersMap := make(map[string]*ContainerInfo, len(pi.IDContainersMap))
	for key, value := range pi.IDContainersMap {
		idContainersMap[key] = value.DeepCopy()
	}
	annotations := make(map[string]string)
	for key, value := range pi.Annotations {
		annotations[key] = value
	}
	return &PodInfo{
		Name:            pi.Name,
		UID:             pi.UID,
		CgroupPath:      pi.CgroupPath,
		Namespace:       pi.Namespace,
		Annotations:     annotations,
		IDContainersMap: idContainersMap,
	}
}
