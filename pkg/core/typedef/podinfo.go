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
	"strings"

	"isula.org/rubik/pkg/core/typedef/cgroup"
)

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
func (pod *PodInfo) DeepCopy() *PodInfo {
	if pod == nil {
		return nil
	}
	contMap := make(map[string]*ContainerInfo, len(pod.IDContainersMap))
	for id, cont := range pod.IDContainersMap {
		contMap[id] = cont.DeepCopy()
	}
	annoMap := make(map[string]string, len(pod.Annotations))
	for k, v := range pod.Annotations {
		annoMap[k] = v
	}
	return &PodInfo{
		Name:            pod.Name,
		UID:             pod.UID,
		CgroupPath:      pod.CgroupPath,
		Namespace:       pod.Namespace,
		Annotations:     annoMap,
		IDContainersMap: contMap,
	}
}

// SetCgroupAttr sets the container cgroup file
func (pod *PodInfo) SetCgroupAttr(key *cgroup.Key, value string) error {
	if err := validateCgroupKey(key); err != nil {
		return err
	}
	return cgroup.WriteCgroupFile(key.SubSys, pod.CgroupPath, key.FileName, value)
}

// GetCgroupAttr gets container cgroup file content
func (pod *PodInfo) GetCgroupAttr(key *cgroup.Key) *cgroup.Attr {
	if err := validateCgroupKey(key); err != nil {
		return &cgroup.Attr{Err: err}
	}
	data, err := cgroup.ReadCgroupFile(key.SubSys, pod.CgroupPath, key.FileName)
	if err != nil {
		return &cgroup.Attr{Err: err}
	}
	return &cgroup.Attr{Value: strings.TrimSpace(string(data)), Err: nil}
}
