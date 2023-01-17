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
// Create: 2023-01-05
// Description: This file contains important interfaces used in the project

// Package api is interface collection
package api

import (
	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/core/typedef"
)

// Registry provides an interface for service discovery
type Registry interface {
	Init() error
	Register(*Service, string) error
	Deregister(*Service, string) error
	GetService(string) (*Service, error)
	ListServices() ([]*Service, error)
}

type ServiceChecker interface {
	Validate() bool
}

type EventFunc interface {
	AddFunc(podInfo *typedef.PodInfo) error
	UpdateFunc(old, new *typedef.PodInfo) error
	DeleteFunc(podInfo *typedef.PodInfo) error
}

// Service contains progress that all services need to have
type Service interface {
	EventFunc
	ServiceChecker
	ID() string
	Setup() error
	TearDown() error
}

type PersistentService interface {
	ServiceChecker
	Setup(viewer Viewer) error
	Start(stopCh <-chan struct{})
}

type ConfigParser interface {
	ParseConfig(data []byte) (map[string]interface{}, error)
	UnmarshalSubConfig(data interface{}, v interface{}) error
}

// Viewer collect on/offline pods info
type Viewer interface {
	ListOnlinePods() ([]*corev1.Pod, error)
	ListOfflinePods() ([]*corev1.Pod, error)
}

// Publisher is a generic interface for Observables
type Publisher interface {
	Subscribe(s Subscriber) error
	Unsubscribe(s Subscriber) error
	Publish(topic typedef.EventType, event typedef.Event)
}

// Subscriber is a common interface for subscribers
type Subscriber interface {
	ID() string
	NotifyFunc(eventType typedef.EventType, event typedef.Event)
	TopicsFunc() []typedef.EventType
}

// EventHandler is the processing interface for change events
type EventHandler interface {
	HandleEvent(eventType typedef.EventType, event typedef.Event)
	EventTypes() []typedef.EventType
}

// Informer is an interface for external pod data sources to interact with rubik
type Informer interface {
	Publisher
	Start(stopCh <-chan struct{})
}
