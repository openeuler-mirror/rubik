//go:build linux
// +build linux

// Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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
	"sync"
	"time"

	"github.com/google/cadvisor/container"
	"github.com/google/cadvisor/manager"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/resource/manager/common"
)

// Manager is the cadvisor manager
type Manager struct {
	manager.Manager
	sync.RWMutex
}

// New creates cadvisor.Manager object
func New(args *Config) *Manager {
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
		log.Errorf("failed to create cadvisor manager: %v", err)
		return nil
	}
	if err := m.Start(); err != nil {
		log.Errorf("failed to start cadvisor manager: %v", err)
		return nil
	}
	return &Manager{
		Manager: m,
	}
}

// Start starts cadvisor manager
func (m *Manager) Start() error {
	m.Lock()
	defer m.Unlock()
	return m.Manager.Start()
}

// Stop stops cadvisor and clear existing factory
func (m *Manager) Stop() error {
	m.Lock()
	defer m.Unlock()
	err := m.Manager.Stop()
	if err != nil {
		return err
	}
	// clear existing factory
	container.ClearContainerHandlerFactories()
	return nil
}

// ContainerInfo gets container infos v2
func (m *Manager) GetPodStats(name string, options common.GetOption) (map[string]common.PodStat, error) {
	m.RLock()
	defer m.RUnlock()
	contInfo, err := m.GetContainerInfoV2(name, options.CadvisorV2RequestOptions)
	if err != nil {
		return nil, err
	}
	var podStats = make(map[string]common.PodStat, len(contInfo))
	for name, info := range contInfo {
		podStats[name] = common.PodStat{
			ContainerInfo: info,
		}
	}
	return podStats, nil
}
