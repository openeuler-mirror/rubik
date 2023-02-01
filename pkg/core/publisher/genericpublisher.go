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
// Description: This file implements pod publisher

// Package publisher implement publisher interface
package publisher

import (
	"fmt"
	"sync"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
)

type (
	// subscriberIDs records subscriber's ID
	subscriberIDs map[string]struct{}
	// NotifyFunc is used to notify subscribers of events
	NotifyFunc func(typedef.EventType, typedef.Event)
)

// defaultPublisher is the default is a globally unique generic publisher entity
var defaultPublisher *genericPublisher

// genericPublisher is the structure to publish Event
type genericPublisher struct {
	sync.RWMutex
	// topicSubscribersMap is a collection of subscribers organized by interested topics
	topicSubscribersMap map[typedef.EventType]subscriberIDs
	// subscribers is the set of notification methods divided by ID
	subscribers map[string]NotifyFunc
}

// newGenericPublisher creates the genericPublisher instance
func newGenericPublisher() *genericPublisher {
	pub := &genericPublisher{
		subscribers:         make(map[string]NotifyFunc, 0),
		topicSubscribersMap: make(map[typedef.EventType]subscriberIDs, 0),
	}
	return pub
}

// getGenericPublisher initializes via lazy mode and return generic publisher entity
func getGenericPublisher() *genericPublisher {
	if defaultPublisher == nil {
		defaultPublisher = newGenericPublisher()
	}
	return defaultPublisher
}

// subscriberExisted confirms the existence of the subscriber based on the ID
func (pub *genericPublisher) subscriberExisted(id string) bool {
	pub.RLock()
	_, ok := pub.subscribers[id]
	pub.RUnlock()
	return ok
}

// Subscribe registers a api.Subscriber
func (pub *genericPublisher) Subscribe(s api.Subscriber) error {
	id := s.ID()
	if pub.subscriberExisted(id) {
		return fmt.Errorf("subscriber %v has registered", id)
	}
	pub.Lock()
	for _, topic := range s.TopicsFunc() {
		if _, ok := pub.topicSubscribersMap[topic]; !ok {
			pub.topicSubscribersMap[topic] = make(subscriberIDs, 0)
		}
		pub.topicSubscribersMap[topic][id] = struct{}{}
		log.Debugf("%s subscribes topic %s", id, topic)
	}
	pub.subscribers[id] = s.NotifyFunc
	pub.Unlock()
	return nil
}

// Unsubscribe unsubscribes the indicated subscriber
func (pub *genericPublisher) Unsubscribe(s api.Subscriber) {
	id := s.ID()
	if !pub.subscriberExisted(id) {
		log.Warnf("subscriber %v has not registered", id)
		return
	}
	pub.Lock()
	for _, topic := range s.TopicsFunc() {
		delete(pub.topicSubscribersMap[topic], id)
		log.Debugf("%s unsubscribes topic %s", id, topic)
	}
	delete(pub.subscribers, id)
	pub.Unlock()
}

// Publish publishes Event to subscribers interested in specified topic
func (pub *genericPublisher) Publish(eventType typedef.EventType, data typedef.Event) {
	log.Debugf("publish %s event", eventType.String())
	pub.RLock()
	for id := range pub.topicSubscribersMap[eventType] {
		pub.subscribers[id](eventType, data)
	}
	pub.RUnlock()
}
