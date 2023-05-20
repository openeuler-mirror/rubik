// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-05-16
// Description: This file is used for metric interface

// Package metric define metric interface
package metric

import "isula.org/rubik/pkg/core/trigger"

// Metric interface defines a series of rubik observation indicator methods
type Metric interface {
	Update() error
	AddTrigger(...trigger.Trigger) Metric
}

// BaseMetric is the basic Metric implementation
type BaseMetric struct {
	Triggers []trigger.Trigger
}

// NewBaseMetric returns a BaseMetric object
func NewBaseMetric() *BaseMetric {
	return &BaseMetric{
		Triggers: make([]trigger.Trigger, 0),
	}
}

// AddTrigger adds trigger methods for metric
func (m *BaseMetric) AddTrigger(triggers ...trigger.Trigger) Metric {
	m.Triggers = append(m.Triggers, triggers...)
	return m
}

// Update updates metric informations
func (m *BaseMetric) Update() error {
	return nil
}
