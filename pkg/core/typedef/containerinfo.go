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
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"isula.org/rubik/pkg/common/util"
)

// ContainerEngineType indicates the type of container engine
type ContainerEngineType int8
type CgroupKey struct {
	SubSys, FileName string
}

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
	Name             string      `json:"name"`
	ID               string      `json:"id"`
	CgroupPath       string      `json:"cgroupPath"`
	RequestResources ResourceMap `json:"requests,omitempty"`
	LimitResources   ResourceMap `json:"limits,omitempty"`
}

// NewContainerInfo creates a ContainerInfo instance
func NewContainerInfo(id, podCgroupPath string, rawContainer *RawContainer) *ContainerInfo {
	requests, limits := rawContainer.GetResourceMaps()
	return &ContainerInfo{
		Name:             rawContainer.status.Name,
		ID:               id,
		CgroupPath:       filepath.Join(podCgroupPath, id),
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
	copyObject.LimitResources = util.DeepCopy(cont.LimitResources).(ResourceMap)
	copyObject.RequestResources = util.DeepCopy(cont.RequestResources).(ResourceMap)
	return &copyObject
}

// SetCgroupAttr sets the container cgroup file
func (cont *ContainerInfo) SetCgroupAttr(key *CgroupKey, value string) error {
	if err := validateCgroupKey(key); err != nil {
		return err
	}
	return util.WriteCgroupFile(key.SubSys, cont.CgroupPath, key.FileName, value)
}

// GetCgroupAttr gets container cgroup file content
func (cont *ContainerInfo) GetCgroupAttr(key *CgroupKey) (string, error) {
	if err := validateCgroupKey(key); err != nil {
		return "", err
	}
	data, err := util.ReadCgroupFile(key.SubSys, cont.CgroupPath, key.FileName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// validateCgroupKey is used to verify the validity of the cgroup key
func validateCgroupKey(key *CgroupKey) error {
	if key == nil {
		return fmt.Errorf("key cannot be empty")
	}
	if len(key.SubSys) == 0 || len(key.FileName) == 0 {
		return fmt.Errorf("invalid key")
	}
	return nil
}
