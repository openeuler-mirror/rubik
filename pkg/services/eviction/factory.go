// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
//	http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-11-04
// Description: This file is used for evict factory

// Package eviction provide cpu eviction services
package eviction

import (
	"sync"

	"isula.org/rubik/pkg/services/eviction/common"
	"isula.org/rubik/pkg/services/eviction/cpu"
	"isula.org/rubik/pkg/services/eviction/memory"
)

const defaultFactoryName string = "EvictFactory"

var (
	// mgr is the singleton manager for reusing resources
	mgr  *common.Manager
	once sync.Once
)

// Factory is the CPUEvict Manager factory class
type Factory struct {
	ObjName string
}

// Name returns the CPUEvict factory class name
func (f Factory) Name() string {
	return defaultFactoryName
}

// NewObj returns a Manager object
func (f Factory) NewObj() (interface{}, error) {
	var e error
	once.Do(
		func() {
			m, err := common.NewManager()
			if err == nil {
				mgr = m
			}
			e = err
		})
	if e == nil && mgr != nil {
		mgr.SetController(f.ObjName, nil)
		switch f.ObjName {
		case common.NodeCPUEvict:
			return cpu.NewManager(mgr, f.ObjName), nil
		case common.NodeMemoryEvict:
			return memory.NewManager(mgr, f.ObjName), nil
		}
	}
	return nil, e
}
