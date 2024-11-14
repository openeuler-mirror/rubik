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
// Description: This file is used for cpu evict service

// Package cpu provide cpu eviction services
package cpu

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/services/eviction/common"
	"isula.org/rubik/pkg/services/helper"
)

// Manager is the cpu manager
type Manager struct {
	*common.Manager
	Name string
}

// NewManager creates the CPU Manager
func NewManager(mgr *common.Manager, name string) *Manager {
	return &Manager{
		Manager: mgr,
		Name:    name,
	}
}

// Run starts the service of cpu manager
func (m *Manager) Run(ctx context.Context) {
	m.Manager.Run(ctx, m.Name)
}

// ID is the name of plugin, must be unique.
func (m *Manager) ID() string {
	return m.Name
}

// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
func (m *Manager) SetConfig(f helper.ConfigHandler) error {
	c, err := fromConfig(m.Name, f)
	if err != nil {
		return fmt.Errorf("failed to create controller %v: %v", m.Name, err)
	}
	m.Manager.SetController(m.Name, c)
	return nil
}

// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
func (m *Manager) GetConfig() interface{} {
	return m.Manager.GetConfig(m.Name)
}
