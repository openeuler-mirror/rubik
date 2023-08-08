// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2023-03-11
// Description: This file is used to implement the factory

// Package helper
package helper

import (
	"errors"
	"fmt"
	"sync"
)

// ServiceFactory is to define Service Factory
type ServiceFactory interface {
	Name() string
	NewObj() (interface{}, error)
}

var (
	rwlock           sync.RWMutex
	serviceFactories = map[string]ServiceFactory{}
)

// AddFactory is to add a service factory
func AddFactory(name string, factory ServiceFactory) error {
	rwlock.Lock()
	defer rwlock.Unlock()
	if _, found := serviceFactories[name]; found {
		return fmt.Errorf("factory already exists")
	}
	serviceFactories[name] = factory
	return nil
}

// GetComponent is to get the interface of object.
func GetComponent(name string) (interface{}, error) {
	rwlock.RLock()
	defer rwlock.RUnlock()
	if f, found := serviceFactories[name]; found {
		return f.NewObj()
	} else {
		return nil, errors.New("factory not found")
	}
}
