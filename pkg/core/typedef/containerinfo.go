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

import (
	"path/filepath"
	"strings"
	"sync"

	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// ContainerEngineType indicates the type of container engine
type ContainerEngineType int8

const (
	// UNDEFINED means undefined container engine
	UNDEFINED ContainerEngineType = iota
	// DOCKER means docker container engine
	DOCKER
	// CONTAINERD means containerd container engine
	CONTAINERD
)

var (
	supportEnginesPrefixMap = map[ContainerEngineType]string{
		DOCKER:     "docker://",
		CONTAINERD: "containerd://",
	}
	currentContainerEngines = UNDEFINED
	setContainerEnginesOnce sync.Once
)

// Support returns true when the container uses the container engine
func (engine *ContainerEngineType) Support(cont *RawContainer) bool {
	if *engine == UNDEFINED {
		return false
	}
	return strings.HasPrefix(cont.status.ContainerID, engine.Prefix())
}

// Prefix returns the ID prefix of the container engine
func (engine *ContainerEngineType) Prefix() string {
	prefix, ok := supportEnginesPrefixMap[*engine]
	if !ok {
		return ""
	}
	return prefix
}

// ContainerInfo contains the interested information of container
type ContainerInfo struct {
	cgroup.Hierarchy
	Name             string      `json:"name"`
	ID               string      `json:"id"`
	RequestResources ResourceMap `json:"requests,omitempty"`
	LimitResources   ResourceMap `json:"limits,omitempty"`
}

// NewContainerInfo creates a ContainerInfo instance
func NewContainerInfo(id, podCgroupPath string, rawContainer *RawContainer) *ContainerInfo {
	requests, limits := rawContainer.GetResourceMaps()
	return &ContainerInfo{
		Name:             rawContainer.status.Name,
		ID:               id,
		Hierarchy:        cgroup.Hierarchy{Path: filepath.Join(podCgroupPath, id)},
		RequestResources: requests,
		LimitResources:   limits,
	}
}

func fixContainerEngine(containerID string) {
	for engine, prefix := range supportEnginesPrefixMap {
		if strings.HasPrefix(containerID, prefix) {
			currentContainerEngines = engine
			return
		}
	}
	currentContainerEngines = UNDEFINED
}

// DeepCopy returns deepcopy object.
func (cont *ContainerInfo) DeepCopy() *ContainerInfo {
	copyObject := *cont
	copyObject.LimitResources = cont.LimitResources.DeepCopy()
	copyObject.RequestResources = cont.RequestResources.DeepCopy()
	return &copyObject
}
