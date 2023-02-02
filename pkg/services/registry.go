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
// Create: 2023-01-28
// Description: This file defines registry for service registration

// Package services implements service registration, discovery and management functions
package services

import (
	"sync"

	"isula.org/rubik/pkg/common/log"
)

type (
	// Creator creates Service objects
	Creator func() interface{}
	// registry is used for service registration
	registry struct {
		sync.RWMutex
		// services is a collection of all registered service
		services map[string]Creator
	}
)

// servicesRegistry  is the globally unique registry
var servicesRegistry = &registry{
	services: make(map[string]Creator, 0),
}

// Register is used to register the service creators
func Register(name string, creator Creator) {
	servicesRegistry.Lock()
	servicesRegistry.services[name] = creator
	servicesRegistry.Unlock()
	log.Debugf("func register (%s)", name)
}

// GetServiceCreator returns the service creator based on the incoming service name
func GetServiceCreator(name string) Creator {
	servicesRegistry.RLock()
	creator, ok := servicesRegistry.services[name]
	servicesRegistry.RUnlock()
	if !ok {
		return nil
	}
	return creator
}
