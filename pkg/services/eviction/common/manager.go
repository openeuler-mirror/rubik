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

// Package eviction provide eviction services
package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	v2 "github.com/google/cadvisor/info/v2"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/metric"
	"isula.org/rubik/pkg/core/trigger/common"
	"isula.org/rubik/pkg/core/trigger/executor"
	"isula.org/rubik/pkg/core/trigger/template"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/resource/analyze"
	"isula.org/rubik/pkg/resource/manager"
	"isula.org/rubik/pkg/resource/manager/cadvisor"
	resource "isula.org/rubik/pkg/resource/manager/common"
	"isula.org/rubik/pkg/services/helper"
)

// evcit type
const (
	NodeCPUEvict    = "cpuevict"
	NodeMemoryEvict = "memoryevict"
)

// requestOptions is the option to get information from cadvisor
var requestOptions = resource.GetOption{
	CadvisorV2RequestOptions: v2.RequestOptions{
		IdType:    v2.TypeName,
		Count:     2,
		Recursive: false,
	},
}

// newCadvisorManager returns cadvisor manager instance
func newCadvisorManager() (resource.Manager, error) {
	const (
		cacheMinutes            = 10
		allowDynamic            = true
		maxHousekeepingInterval = 10 * time.Second
		metricSet               = "cpu,memory"
	)
	return manager.GetManagerBuilder(manager.CADVISOR)(
		cadvisor.NewConfig(
			cadvisor.WithCacheAge(cacheMinutes*time.Minute),
			cadvisor.WithMetrics(metricSet),
			cadvisor.WithHouseKeepingInterval(maxHousekeepingInterval),
			cadvisor.WithHouseKeepingDynamic(allowDynamic),
		))
}

// Controller is a controller for different resources
type Controller interface {
	Start(context.Context, func() error)
	Config() interface{}
}

// Manager is used to manage Evcit services
type Manager struct {
	sync.RWMutex
	helper.ServiceBase
	viewer          api.Viewer
	analyzer        *analyze.Analyzer // analyzer is used to assist in analyzing Pod data
	analyzerStarted bool
	baseMetric      *metric.BaseMetric // met is used to implement condition-triggered
	controllers     map[string]Controller
}

// NewManager returns a instance of evict manager
func NewManager() (*Manager, error) {
	// 1. Use cadvisor as a manager of Pod data
	cm, err := newCadvisorManager()
	if err != nil {
		return nil, err
	}

	// 2. Analyze Pod resources through cadvisor data
	analyzer := analyze.NewResourceAnalyzer(cm)

	// 3. Define different kinds of triggers, including sorting, eviction
	var (
		evictTrigger = template.FromBaseTemplate(
			template.WithName("eviction"),
			template.WithPodAction(executor.EvictPod),
		)
		cpuTrigger = template.FromBaseTemplate(
			template.WithName("node_cpu_trigger"),
			template.WithPodTransformation(executor.MaxValueTransformer(analyzer.CPUCalculatorBuilder(&requestOptions))),
		).SetNext(evictTrigger)
		memoryTrigger = template.FromBaseTemplate(
			template.WithName("node_memory_trigger"),
			template.WithPodTransformation(executor.MaxValueTransformer(analyzer.MemoryCalculatorBuilder(&requestOptions))),
		).SetNext(evictTrigger)
	)

	return &Manager{
		ServiceBase: helper.ServiceBase{
			Name: "eviction",
		},
		analyzer:    analyzer,
		controllers: make(map[string]Controller),
		baseMetric: &metric.BaseMetric{
			Triggers: map[string][]common.Trigger{
				NodeCPUEvict: {
					cpuTrigger,
				},
				NodeMemoryEvict: {
					memoryTrigger,
				},
			},
		}}, nil
}

// SetController sets the resource controller for the manager
func (m *Manager) SetController(name string, c Controller) {
	m.Lock()
	m.controllers[name] = c
	m.Unlock()
}

// GetConfig returns the config
func (m *Manager) GetConfig(name string) interface{} {
	m.Lock()
	defer m.Unlock()
	if c, existed := m.controllers[name]; existed {
		return c.Config()
	}
	return nil
}

// Run checks resources cyclically.
func (m *Manager) Run(ctx context.Context, name string) {
	// 1. start Resource Manager to collect data
	m.Lock()
	if !m.analyzerStarted {
		m.analyzer.Start()
		m.analyzerStarted = true
	}
	// 2. start diverse collectors
	controller, existed := m.controllers[name]
	m.Unlock()
	if !existed {
		log.Errorf("failed to find controller %v", name)
		return
	}
	controller.Start(ctx, m.alarm(name))
}

// IsRunner returns true that tells other Manager is a persistent service
func (m *Manager) IsRunner() bool {
	return true
}

// PreStart is the pre-start action
func (m *Manager) PreStart(viewer api.Viewer) error {
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	m.viewer = viewer
	return nil
}

func priority(online bool) api.ListOption {
	const (
		offTag = "true"
		onTag  = "false"
	)
	var wanted = onTag
	if !online {
		wanted = offTag
	}
	return func(pod *typedef.PodInfo) bool {
		if prio, declared := pod.Annotations[constant.PriorityAnnotationKey]; declared {
			return prio == wanted
		}
		// If no priority is set, the priority is online by default
		return onTag == wanted
	}
}

// Terminate clean the resource
func (m *Manager) Terminate(api.Viewer) error {
	var err error
	m.Lock()
	if m.analyzer != nil {
		err = m.analyzer.Stop()
	}
	m.analyzer = nil
	m.Unlock()
	return err
}

func (m *Manager) alarm(typ string) func() error {
	return func() error {
		var (
			errs error
			ctx  = context.WithValue(context.Background(), common.TARGETPODS, m.viewer.ListPodsWithOptions(priority(false)))
		)
		for _, t := range m.baseMetric.Triggers[typ] {
			errs = util.AppendErr(errs, t.Activate(ctx))
		}
		return errs
	}
}
