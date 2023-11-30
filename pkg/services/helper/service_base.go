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
// Description: This file is the base of service.

// Package helper provide some helper for service.
package helper

import (
	"context"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
)

// ServiceBase is the basic class of a service.
type ServiceBase struct {
	Name string
}

// NewServiceBase returns the instance of service
func NewServiceBase(serviceName string) *ServiceBase {
	return &ServiceBase{
		Name: serviceName,
	}
}

// ConfigHandler is that obtains the configured callback function.
type ConfigHandler func(configName string, d interface{}) error

// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
func (s *ServiceBase) SetConfig(ConfigHandler) error {
	return nil
}

// PreStart is an interface for calling a collection of methods when the service is pre-started
func (s *ServiceBase) PreStart(api.Viewer) error {
	log.Warnf("%v: PreStart interface is not implemented", s.Name)
	return nil
}

// Terminate is an interface that calls a collection of methods when the service terminates
func (s *ServiceBase) Terminate(api.Viewer) error {
	log.Warnf("%v: Terminate interface is not implemented", s.Name)
	return nil
}

// ID is an interface that calls a collection of methods returning service's ID
func (s *ServiceBase) ID() string {
	return s.Name
}

// IsRunner to Confirm whether it is a runner
func (s *ServiceBase) IsRunner() bool {
	return false
}

// Run to start runner
func (s *ServiceBase) Run(context.Context) {}

// Stop to stop runner
func (s *ServiceBase) Stop() error {
	return nil
}

// AddPod to deal the event of adding a pod.
func (s *ServiceBase) AddPod(*typedef.PodInfo) error {
	return nil
}

// UpdatePod to deal the pod update event.
func (s *ServiceBase) UpdatePod(old, new *typedef.PodInfo) error {
	return nil
}

// DeletePod to deal the pod deletion event.
func (s *ServiceBase) DeletePod(*typedef.PodInfo) error {
	return nil
}

// GetConfig returns the config of service
func (s *ServiceBase) GetConfig() interface{} {
	return nil
}
