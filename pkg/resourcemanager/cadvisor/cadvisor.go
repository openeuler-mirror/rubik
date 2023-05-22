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

	cmemory "github.com/google/cadvisor/cache/memory"
	cadvisorcontainer "github.com/google/cadvisor/container"
	"github.com/google/cadvisor/manager"
	csysfs "github.com/google/cadvisor/utils/sysfs"
	"k8s.io/klog/v2"
)

type CadvisorManager struct {
	cgroupDriver string
	manager.Manager
}

func NewCadvidorManager() *CadvisorManager {
	var includedMetrics = cadvisorcontainer.MetricSet{
		cadvisorcontainer.CpuUsageMetrics:         struct{}{},
		cadvisorcontainer.ProcessSchedulerMetrics: struct{}{},
	}

	allowDynamic := true
	maxHousekeepingInterval := 10 * time.Second
	memCache := cmemory.New(10*time.Minute, nil)
	sysfs := csysfs.NewRealSysFs()
	maxHousekeepingConfig := manager.HouskeepingConfig{Interval: &maxHousekeepingInterval, AllowDynamic: &allowDynamic}

	m, err := manager.New(memCache, sysfs, maxHousekeepingConfig, includedMetrics, http.DefaultClient, []string{"/kubepods"}, []string{""}, "", time.Duration(1))
	if err != nil {
		klog.Errorf("Failed to create cadvisor manager start: %v", err)
		return nil
	}

	if err := m.Start(); err != nil {
		klog.Errorf("Failed to start cadvisor manager: %v", err)
		return nil
	}

	return &CadvisorManager{
		Manager: m,
	}
}
