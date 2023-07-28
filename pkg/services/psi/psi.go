// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
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
// Date: 2023-05-16
// Description: This file is used for psi service
package psi

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/metric"
	"isula.org/rubik/pkg/core/trigger"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services/helper"
)

const (
	minInterval          = 10
	maxInterval          = 30
	maxThreshold float64 = 100.0
	minThreshold float64 = 5.0
)

// Factory is the PSI Manager factory class
type Factory struct {
	ObjName string
}

// Name returns the factory class name
func (f Factory) Name() string {
	return "PSIFactory"
}

// NewObj returns a Manager object
func (f Factory) NewObj() (interface{}, error) {
	return NewManager(f.ObjName), nil
}

// Config is PSI service configuration
type Config struct {
	Interval       int      `json:"interval,omitempty"`
	Avg10Threshold float64  `json:"avg10threshold,omitempty"`
	Resource       []string `json:"resource,omitempty"`
}

// NewConfig returns default psi configuration
func NewConfig() *Config {
	return &Config{
		Interval:       minInterval,
		Resource:       make([]string, 0),
		Avg10Threshold: defaultAvg10Threshold,
	}
}

// Validate verifies that the psi parameter is set correctly
func (conf *Config) Validate() error {
	if conf.Interval < minInterval || conf.Interval > maxInterval {
		return fmt.Errorf("interval should in the range [%v, %v]", minInterval, maxInterval)
	}
	if conf.Avg10Threshold < minThreshold || conf.Avg10Threshold > maxThreshold {
		return fmt.Errorf("avg10 threshold should in the range [%v, %v]", minThreshold, maxThreshold)
	}
	if len(conf.Resource) == 0 {
		return fmt.Errorf("specify at least one type resource")
	}
	for _, res := range conf.Resource {
		if _, support := supportResources[res]; !support {
			return fmt.Errorf("%v type resource is not supported", res)
		}
	}
	return nil
}

// Manager is used to manage PSI services
type Manager struct {
	conf   *Config
	Viewer api.Viewer
	helper.ServiceBase
}

// NewManager returns psi manager objects
func NewManager(n string) *Manager {
	return &Manager{
		ServiceBase: helper.ServiceBase{
			Name: n,
		},
		conf: NewConfig(),
	}
}

// Run checks psi metrics cyclically.
func (m *Manager) Run(ctx context.Context) {
	wait.Until(
		func() {
			if err := m.monitor(); err != nil {
				log.Errorf("failed to monitor PSI metrics: %v", err)
			}
		},
		time.Second*time.Duration(m.conf.Interval),
		ctx.Done())
}

// SetConfig sets and checks Config
func (m *Manager) SetConfig(f helper.ConfigHandler) error {
	var conf = NewConfig()
	if err := f(m.Name, conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}
	m.conf = conf
	return nil
}

// IsRunner returns true that tells other Manager is a persistent service
func (m *Manager) IsRunner() bool {
	return true
}

// PreStart is the pre-start action
func (m *Manager) PreStart(viewer api.Viewer) error {
	m.Viewer = viewer
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

// monitor gets metrics and fire triggers when satisfied
func (m *Manager) monitor() error {
	metric := &BasePSIMetric{conservation: m.Viewer.ListPodsWithOptions(priority(true)),
		suspicion:      m.Viewer.ListPodsWithOptions(priority(false)),
		BaseMetric:     metric.NewBaseMetric(),
		resources:      m.conf.Resource,
		avg10Threshold: m.conf.Avg10Threshold,
	}
	metric.AddTrigger(
		trigger.NewTrigger(trigger.RESOURCEANALYSIS).
			SetNext(trigger.NewTrigger(trigger.EXPULSION)),
	)
	if err := metric.Update(); err != nil {
		return err
	}
	return nil
}
