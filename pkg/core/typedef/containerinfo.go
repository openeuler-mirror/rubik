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
// Description: This file defines ContainerInfo

// Package typedef defines core struct and methods for rubik
package typedef

import "path/filepath"

// ContainerInfo contains the interested information of container
type ContainerInfo struct {
	// Basic Information
	Name       string `json:"name"`
	ID         string `json:"id"`
	PodID      string `json:"podID"`
	CgroupRoot string `json:"cgroupRoot"`
	CgroupAddr string `json:"cgroupAddr"`
}

// NewContainerInfo creates a ContainerInfo instance
func NewContainerInfo(container RawContainer, podID, conID, cgroupRoot, podCgroupPath string) *ContainerInfo {
	c := ContainerInfo{
		Name:       container.Name,
		ID:         conID,
		PodID:      podID,
		CgroupRoot: cgroupRoot,
		CgroupAddr: filepath.Join(podCgroupPath, conID),
	}
	return &c
}

// CgroupPath returns cgroup path of specified subsystem of a container
func (ci *ContainerInfo) CgroupPath(subsys string) string {
	if ci == nil || ci.Name == "" {
		return ""
	}
	return filepath.Join(ci.CgroupRoot, subsys, ci.CgroupAddr)
}

// Clone returns deepcopy object.
func (ci *ContainerInfo) Clone() *ContainerInfo {
	copy := *ci
	return &copy
}
