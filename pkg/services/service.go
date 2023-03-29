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
// Description: This file is the Interface set of services

// Package services
package services

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services/helper"
)

// PodEvent for listening to pod changes.
type PodEvent interface {
	// Deal processing adding a pod.
	AddPod(*typedef.PodInfo) error
	// Deal processing update a pod config.
	UpdatePod(old, new *typedef.PodInfo) error
	// Deal processing delete a pod.
	DeletePod(*typedef.PodInfo) error
}

// Runner for background service process.
type Runner interface {
	// IsRunner for Confirm whether it is
	IsRunner() bool
	// Start runner
	Run(context.Context)
	// Stop runner
	Stop() error
}

// Service interface contains methods which must be implemented by all services.
type Service interface {
	Runner
	PodEvent
	// ID is the name of plugin, must be unique.
	ID() string
	// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
	SetConfig(helper.ConfigHandler) error
	// GetConfig is an interface for obtaining service running configurations.
	GetConfig() interface{}
	// PreStarter is an interface for calling a collection of methods when the service is pre-started
	PreStart(api.Viewer) error
	// Terminator is an interface that calls a collection of methods when the service terminates
	// it will stop runner and clear configuration
	Terminate(api.Viewer) error
}

// FeatureSpec to defines the feature name and whether the feature is enabled.
type FeatureSpec struct {
	// feature name
	Name string
	// Default is the default enablement state for the feature
	Default bool
}

// InitServiceComponents for initilize serverice components
func InitServiceComponents(specs []FeatureSpec) {
	for _, spec := range specs {
		if !spec.Default {
			log.Warnf("feature is disabled by default:%v", spec.Name)
			continue
		}

		initFunc, found := serviceComponents[spec.Name]
		if !found {
			log.Errorf("init service failed, name:%v", spec.Name)
			continue
		}

		if err := initFunc(spec.Name); err != nil {
			log.Warnf("init component failed, name:%v,error:%v", spec.Name, err)
		}
	}
}

// GetServiceComponent to get the component service interface.
func GetServiceComponent(name string) (Service, error) {
	si, err := helper.GetComponent(name)
	if err != nil {
		return nil, fmt.Errorf("get service failed, name:%v,err:%v", name, err)
	}
	srv, ok := si.(Service)
	if !ok || srv == nil {
		return nil, fmt.Errorf("failed to convert the type,name:%v", name)
	}
	return srv, nil
}
