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

	"isula.org/rubik/pkg/common/log"
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
	Name       string `json:"name"`
	ID         string `json:"id"`
	CgroupPath string `json:"cgroupPath"`
}

// NewContainerInfo creates a ContainerInfo instance
func NewContainerInfo(name, id, podCgroupPath string) *ContainerInfo {
	return &ContainerInfo{
		Name:       name,
		ID:         id,
		CgroupPath: filepath.Join(podCgroupPath, id),
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

// getRealContainerID parses the containerID of k8s
func (cont *RawContainer) getRealContainerID() string {
	/*
		Note:
		An UNDEFINED container engine was used when the function was executed for the first time
		it seems unlikely to support different container engines at runtime,
		So we don't consider the case of midway container engine changes
		`fixContainerEngine` is only executed when `getRealContainerID` is called for the first time
	*/
	setContainerEnginesOnce.Do(func() { fixContainerEngine(cont.status.ContainerID) })

	if !currentContainerEngines.Support(cont) {
		log.Errorf("fatal error : unsupported container engine")
		return ""
	}

	cid := cont.status.ContainerID[len(currentContainerEngines.Prefix()):]
	// the container may be in the creation or deletion phase.
	if len(cid) == 0 {
		return ""
	}
	return cid
}

// DeepCopy returns deepcopy object.
func (info *ContainerInfo) DeepCopy() *ContainerInfo {
	copyObject := *info
	return &copyObject
}
