// Copyright (c) Huawei Technologies Co., Ltd. 2021-2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-09-07
// Description: This file is used for container engines

package typedef

import (
	"strings"
	"sync"
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
	// ISULAD means isulad container engine
	ISULAD
	// CRIO means crio container engine
	CRIO
)

var (
	supportEnginesPrefixMap = map[ContainerEngineType]string{
		DOCKER:     "docker://",
		CONTAINERD: "containerd://",
		ISULAD:     "iSulad://",
		CRIO:       "cri-o://",
	}
	containerEngineScopes = map[ContainerEngineType]string{
		DOCKER:     "docker",
		CONTAINERD: "cri-containerd",
		ISULAD:     "isulad",
		CRIO:       "crio",
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
