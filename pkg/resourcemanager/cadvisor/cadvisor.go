//go:build linux
// +build linux

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
// Create: 2023-05-21
// Description: This file defines cadvisor

package cadvisor

import (
	"net/http"
	"time"

	"github.com/google/cadvisor/cache/memory"
	"github.com/google/cadvisor/container"
	v2 "github.com/google/cadvisor/info/v2"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/utils/sysfs"

	"isula.org/rubik/pkg/common/log"
)

// Manager is the cadvisor manager
type Manager struct {
	manager.Manager
}

// StartArgs is a set of parameters that control the startup of cadvisor
type StartArgs struct {
	MemCache              *memory.InMemoryCache
	SysFs                 sysfs.SysFs
	IncludeMetrics        container.MetricSet
	MaxHousekeepingConfig manager.HouskeepingConfig
}

// WithStartArgs creates cadvisor.Manager object
func WithStartArgs(args StartArgs) *Manager {
	const (
		perfEventsFile = "/sys/kernel/debug/tracing/events/raw_syscalls/sys_enter"
	)
	var (
		rawContainerCgroupPathPrefixWhiteList = []string{"/kubepods"}
		containerEnvMetadataWhiteList         = []string{}
		resctrlInterval                       = time.Second
	)
	m, err := manager.New(args.MemCache, args.SysFs, args.MaxHousekeepingConfig,
		args.IncludeMetrics, http.DefaultClient, rawContainerCgroupPathPrefixWhiteList,
		containerEnvMetadataWhiteList, perfEventsFile, resctrlInterval)
	if err != nil {
		log.Errorf("Failed to create cadvisor manager: %v", err)
		return nil
	}
	if err := m.Start(); err != nil {
		log.Errorf("Failed to start cadvisor manager: %v", err)
		return nil
	}
	return &Manager{
		Manager: m,
	}
}

// Start starts cadvisor manager
func (c *Manager) Start() error {
	return c.Manager.Start()
}

// Stop stops cadvisor and clear existing factory
func (c *Manager) Stop() error {
	err := c.Manager.Stop()
	if err != nil {
		return err
	}
	// clear existing factory
	container.ClearContainerHandlerFactories()
	return nil
}

// ContainerInfo gets container infos v2
func (c *Manager) ContainerInfoV2(name string,
	options v2.RequestOptions) (map[string]v2.ContainerInfo, error) {
	return c.GetContainerInfoV2(name, options)
}
