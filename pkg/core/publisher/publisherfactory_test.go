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
// Description: This file tests publisher factory

// Package publisher implement publisher interface
package publisher

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetPublisherFactory tests GetPublisherFactory
func TestGetPublisherFactory(t *testing.T) {
	tests := []struct {
		name string
		want *Factory
	}{
		{
			name: "TC1-success",
			want: &Factory{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPublisherFactory()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublisherFactory() = %v, want %v", got, tt.want)
			}
			assert.NotNil(t, got.GetPublisher(GENERIC))
			const typ publihserType = 100
			assert.Nil(t, got.GetPublisher(typ))
		})
	}
}
