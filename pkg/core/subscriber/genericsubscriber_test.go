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
// Description: This file tests the generic subscriber functionality

// Package subscriber implements generic subscriber interface
package subscriber

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
)

type mockEventHandler struct{}

// HandleEvent handles the event from publisher
func (h *mockEventHandler) HandleEvent(eventType typedef.EventType, event typedef.Event) {}

// EventTypes returns the intersted event types
func (h *mockEventHandler) EventTypes() []typedef.EventType {
	return nil
}

// TestNewGenericSubscriber tests NewGenericSubscriber
func TestNewGenericSubscriber(t *testing.T) {
	type args struct {
		handler api.EventHandler
		id      string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TC1-NewGenericSubscriber/ID/NotifyFunc/TopicsFunc",
			args: args{
				handler: &mockEventHandler{},
				id:      "rubik",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGenericSubscriber(tt.args.handler, tt.args.id)
			assert.Equal(t, "rubik", got.ID())
			got.NotifyFunc(typedef.INFOADD, nil)
			got.TopicsFunc()
		})
	}
}
