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
// Description: This file implements the generic subscriber functionality

// Package subscriber implements generic subscriber interface
package subscriber

import (
	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
)

type genericSubscriber struct {
	id string
	api.EventHandler
}

// NewGenericSubscriber returns the generic subscriber entity
func NewGenericSubscriber(handler api.EventHandler, id string) *genericSubscriber {
	return &genericSubscriber{
		id:           id,
		EventHandler: handler,
	}
}

// ID returns the unique ID of the subscriber
func (pub *genericSubscriber) ID() string {
	return pub.id
}

// NotifyFunc notifys subscriber event
func (pub *genericSubscriber) NotifyFunc(eventType typedef.EventType, event typedef.Event) {
	pub.HandleEvent(eventType, event)
}

// TopicsFunc returns the topics that the subscriber is interested in
func (pub *genericSubscriber) TopicsFunc() []typedef.EventType {
	return pub.EventTypes()
}
