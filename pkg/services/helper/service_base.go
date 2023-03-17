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

// Package helper
package helper

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
)

// ServiceBase is the basic class of a service.
type ServiceBase struct{}

// ConfigHandler is that obtains the configured callback function.
type ConfigHandler func(configName string, d interface{}) error

// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
func (s *ServiceBase) SetConfig(ConfigHandler) error {
	return nil
}

// PreStart is an interface for calling a collection of methods when the service is pre-started
func (s *ServiceBase) PreStart(api.Viewer) error {
	log.Warnf("this interface is not implemented.")
	return nil
}

// Terminate is an interface that calls a collection of methods when the service terminates
func (s *ServiceBase) Terminate(api.Viewer) error {
	log.Warnf("this interface is not implemented.")
	return nil
}

// IsRunner to Confirm whether it is a runner
func (s *ServiceBase) IsRunner() bool {
	return false
}

// Run to start runner
func (s *ServiceBase) Run(context.Context) {}

// Stop to stop runner
func (s *ServiceBase) Stop() error {
	return fmt.Errorf("this interface is not implemented")
}

// AddPod to deal the event of adding a pod.
func (s *ServiceBase) AddPod(*typedef.PodInfo) error {
	log.Warnf("this interface is not implemented.")
	return nil
}

// UpdatePod to deal the pod update event.
func (S *ServiceBase) UpdatePod(old, new *typedef.PodInfo) error {
	log.Warnf("this interface is not implemented.")
	return nil
}

// DeletePod to deal the pod deletion event.
func (s *ServiceBase) DeletePod(*typedef.PodInfo) error {
	log.Warnf("this interface is not implemented.")
	return nil
}

// Run to start runner
func (s *ServiceBase) GetConfig() interface{} {
	return nil
}
