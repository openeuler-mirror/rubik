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

// ListOption is for filtering podInfo
type ListOption func(pi *typedef.PodInfo) bool

// Viewer collect on/offline pods info
type Viewer interface {
	ListContainersWithOptions(options ...ListOption) map[string]*typedef.ContainerInfo
	ListPodsWithOptions(options ...ListOption) map[string]*typedef.PodInfo
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
	WaitReady()
}
