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
// Create: 2023-03-23
// Description: This file tests pod publisher

// Package publisher implement publisher interface
package publisher

import (
	"testing"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
)

// mockSubscriber is used to mock a subscriber
type mockSubscriber struct {
	name string
}

//  ID returns the unique ID of the subscriber
func (s *mockSubscriber) ID() string {
	return s.name
}

// NotifyFunc notifys subscriber event
func (s *mockSubscriber) NotifyFunc(eventType typedef.EventType, event typedef.Event) {}

// TopicsFunc returns the topics that the subscriber is interested in
func (s *mockSubscriber) TopicsFunc() []typedef.EventType {
	return []typedef.EventType{typedef.RAWPODADD}
}

// Test_genericPublisher_Subscribe tests Subscribe of genericPublisher
func Test_genericPublisher_Subscribe(t *testing.T) {
	const subID = "ID"
	type fields struct {
		topicSubscribersMap map[typedef.EventType]subscriberIDs
		subscribers         map[string]NotifyFunc
	}
	type args struct {
		s api.Subscriber
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TC1-subscriber existed",
			fields: fields{
				topicSubscribersMap: map[typedef.EventType]subscriberIDs{
					typedef.INFOADD: map[string]struct{}{subID: {}},
				},
				subscribers: map[string]NotifyFunc{subID: nil},
			},
			args: args{
				s: &mockSubscriber{name: subID},
			},
			wantErr: true,
		},
		{
			name: "TC2-subscriber is not existed",
			fields: fields{
				topicSubscribersMap: make(map[typedef.EventType]subscriberIDs),
				subscribers:         make(map[string]NotifyFunc),
			},
			args: args{
				s: &mockSubscriber{name: subID},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub := newGenericPublisher()
			pub.subscribers = tt.fields.subscribers
			pub.topicSubscribersMap = tt.fields.topicSubscribersMap
			if err := pub.Subscribe(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("genericPublisher.Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
			pub.Publish(typedef.RAWPODADD, "a")
		})
	}
}

// Test_genericPublisher_Unsubscribe tests Unsubscribe of genericPublisher
func Test_genericPublisher_Unsubscribe(t *testing.T) {
	const subID = "ID"
	type fields struct {
		topicSubscribersMap map[typedef.EventType]subscriberIDs
		subscribers         map[string]NotifyFunc
	}
	type args struct {
		s api.Subscriber
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TC1-subscriber existed",
			fields: fields{
				topicSubscribersMap: map[typedef.EventType]subscriberIDs{
					typedef.INFOADD: map[string]struct{}{subID: {}},
				},
				subscribers: map[string]NotifyFunc{subID: nil},
			},
			args: args{
				s: &mockSubscriber{name: subID},
			},
		},
		{
			name: "TC2-subscriber is not existed",
			fields: fields{
				topicSubscribersMap: make(map[typedef.EventType]subscriberIDs),
				subscribers:         make(map[string]NotifyFunc),
			},
			args: args{
				s: &mockSubscriber{name: subID},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub := &genericPublisher{
				topicSubscribersMap: tt.fields.topicSubscribersMap,
				subscribers:         tt.fields.subscribers,
			}
			pub.Unsubscribe(tt.args.s)
		})
	}
}
