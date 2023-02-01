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
	"context"

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

type ServiceDescriber interface {
	ID() string
}

type EventFunc interface {
	AddFunc(podInfo *typedef.PodInfo) error
	UpdateFunc(old, new *typedef.PodInfo) error
	DeleteFunc(podInfo *typedef.PodInfo) error
}

// Service contains progress that all services need to have
type Service interface {
	ServiceDescriber
	EventFunc
}

// PersistentService is an abstract persistent running service
type PersistentService interface {
	ServiceDescriber
	// Run is a service processing logic, which is blocking (implemented in an infinite loop, etc.)
	Run(ctx context.Context)
}

type ConfigParser interface {
	ParseConfig(data []byte) (map[string]interface{}, error)
	UnmarshalSubConfig(data interface{}, v interface{}) error
}

// Viewer collect on/offline pods info
type Viewer interface {
	ListOnlinePods() ([]*typedef.PodInfo, error)
	ListOfflinePods() ([]*typedef.PodInfo, error)
}

// Publisher is a generic interface for Observables
type Publisher interface {
	Subscribe(s Subscriber) error
	Unsubscribe(s Subscriber)
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
	Start(ctx context.Context)
}

// Logger is the handler to print the log
type Logger interface {
	// Errorf logs bugs that affect normal functionality
	Errorf(f string, args ...interface{})
	// Warnf logs produce unexpected results
	Warnf(f string, args ...interface{})
	// Infof logs normal messages
	Infof(f string, args ...interface{})
	// Debugf logs verbose messages
	Debugf(f string, args ...interface{})
}
