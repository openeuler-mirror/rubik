// Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-02-17
// Description: This file is used for testing podmanager

package podmanager

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
)

func TestPodManager_ListContainersWithOptions(t *testing.T) {
	var (
		cont1 = &typedef.ContainerInfo{
			ID: "testCon1",
		}
		cont2 = &typedef.ContainerInfo{
			ID: "testCon2",
		}
		cont3 = &typedef.ContainerInfo{
			ID: "testCon3",
		}
	)

	type fields struct {
		pods *PodCache
	}
	type args struct {
		options []api.ListOption
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*typedef.ContainerInfo
	}{
		{
			name: "TC1-filter priority container",
			args: args{
				[]api.ListOption{
					func(pi *typedef.PodInfo) bool {
						return pi.Annotations[constant.PriorityAnnotationKey] == "true"
					},
				},
			},
			fields: fields{
				pods: &PodCache{
					Pods: map[string]*typedef.PodInfo{
						"testPod1": {
							UID: "testPod1",
							IDContainersMap: map[string]*typedef.ContainerInfo{
								cont1.ID: cont1,
								cont2.ID: cont2,
							},
							Annotations: map[string]string{
								constant.PriorityAnnotationKey: "true",
							},
						},
						"testPod2": {
							IDContainersMap: map[string]*typedef.ContainerInfo{
								cont3.ID: cont3,
							},
						},
					},
				},
			},
			want: map[string]*typedef.ContainerInfo{
				cont1.ID: cont1,
				cont2.ID: cont2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &PodManager{
				Pods: tt.fields.pods,
			}
			if got := manager.ListContainersWithOptions(tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, tt.want, got)
				t.Errorf("PodManager.ListContainersWithOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
