// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2023-02-10
// Description: This file test qos level setting service

// Package qos is the service used for qos level setting
package preemption

import (
	"testing"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/tests/try"
)

func init() {
	try.InitTestCGRoot(try.TestRoot)
}

type fields struct {
	Name   string
	Config PreemptionConfig
}
type args struct {
	old *try.FakePod
	new *try.FakePod
}

type test struct {
	name    string
	fields  fields
	args    args
	wantErr bool
	preHook func(*try.FakePod) *try.FakePod
}

var getCommonField = func(subSys []string) fields {
	return fields{
		Name:   "qos",
		Config: PreemptionConfig{Resource: subSys},
	}
}

// TestPreemptionAddFunc tests AddFunc of Preemption
func TestPreemptionAddFunc(t *testing.T) {
	const containerNum = 3
	var addFuncTC = []test{
		{
			name:   "TC1-set offline pod qos ok",
			fields: getCommonField([]string{"cpu", "memory"}),
			args: args{
				new: try.GenFakeOfflinePod(map[*cgroup.Key]string{
					supportCgroupTypes["cpu"].cgKey:    "0",
					supportCgroupTypes["memory"].cgKey: "0",
				}),
			},
		},
		{
			name:   "TC2-set online pod qos ok",
			fields: getCommonField([]string{"cpu", "memory"}),
			args: args{
				new: try.GenFakeOnlinePod(map[*cgroup.Key]string{
					supportCgroupTypes["cpu"].cgKey:    "0",
					supportCgroupTypes["memory"].cgKey: "0",
				}).WithContainers(containerNum),
			},
		},
		{
			name:    "TC3-empty pod info",
			fields:  getCommonField([]string{"cpu", "memory", "net"}),
			wantErr: true,
		},
		{
			name:   "TC4-invalid annotation key",
			fields: getCommonField([]string{"cpu"}),
			args: args{
				new: try.GenFakeBestEffortPod(map[*cgroup.Key]string{supportCgroupTypes["cpu"].cgKey: "0"}),
			},
			preHook: func(pod *try.FakePod) *try.FakePod {
				newPod := pod.DeepCopy()
				newPod.Annotations["undefine"] = "true"
				return newPod
			},
		},
		{
			name:   "TC5-invalid annotation value",
			fields: getCommonField([]string{"cpu"}),
			args: args{
				new: try.GenFakeBestEffortPod(map[*cgroup.Key]string{supportCgroupTypes["cpu"].cgKey: "0"}),
			},
			preHook: func(pod *try.FakePod) *try.FakePod {
				newPod := pod.DeepCopy()
				newPod.Annotations[constant.PriorityAnnotationKey] = "undefine"
				return newPod
			},
		},
	}

	for _, tt := range addFuncTC {
		t.Run(tt.name, func(t *testing.T) {
			q := &Preemption{
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
				config: tt.fields.Config,
			}
			if tt.preHook != nil {
				tt.preHook(tt.args.new)
			}
			if tt.args.new != nil {
				if err := q.AddPod(tt.args.new.PodInfo); (err != nil) != tt.wantErr {
					t.Errorf("QoS.AddPod() error = %v, wantErr %v", err, tt.wantErr)
				}

			}
			tt.args.new.CleanPath().OrDie()
		})
	}
}

// TestPreemptionUpdatePod tests UpdatePod of Preemption
func TestPreemptionUpdatePod(t *testing.T) {
	var updateFuncTC = []test{
		{
			name:   "TC1-online to offline",
			fields: getCommonField([]string{"cpu"}),
			args:   args{old: try.GenFakeOnlinePod(map[*cgroup.Key]string{supportCgroupTypes["cpu"].cgKey: "0"}).WithContainers(3)},
			preHook: func(pod *try.FakePod) *try.FakePod {
				newPod := pod.DeepCopy()
				newAnnotation := make(map[string]string, 0)
				newAnnotation[constant.PriorityAnnotationKey] = "true"
				newPod.Annotations = newAnnotation
				return newPod
			},
		},
		{
			name:   "TC2-offline to online",
			fields: getCommonField([]string{"cpu"}),
			args:   args{old: try.GenFakeOfflinePod(map[*cgroup.Key]string{supportCgroupTypes["cpu"].cgKey: "0"})},
			preHook: func(pod *try.FakePod) *try.FakePod {
				newPod := pod.DeepCopy()
				newAnnotation := make(map[string]string, 0)
				newAnnotation[constant.PriorityAnnotationKey] = "false"
				newPod.Annotations = newAnnotation
				return newPod
			},
			wantErr: true,
		},
		{
			name:   "TC3-online to online",
			fields: getCommonField([]string{"cpu"}),
			args:   args{old: try.GenFakeOnlinePod(map[*cgroup.Key]string{supportCgroupTypes["cpu"].cgKey: "0"})},
			preHook: func(pod *try.FakePod) *try.FakePod {
				return pod.DeepCopy()
			},
		},
	}

	for _, tt := range updateFuncTC {
		t.Run(tt.name, func(t *testing.T) {
			q := &Preemption{
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
				config: tt.fields.Config,
			}
			if tt.preHook != nil {
				tt.args.new = tt.preHook(tt.args.old)
			}
			if err := q.UpdatePod(tt.args.old.PodInfo, tt.args.new.PodInfo); (err != nil) != tt.wantErr {
				t.Errorf("QoS.UpdatePod() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.new.CleanPath().OrDie()
			tt.args.old.CleanPath().OrDie()
		})
	}
}

func TestPreemptionValidate(t *testing.T) {
	var validateTC = []test{
		{
			name: "TC1-normal config",
			fields: fields{
				Name:   "qos",
				Config: PreemptionConfig{Resource: []string{"cpu", "memory"}},
			},
		},
		{
			name: "TC2-abnormal config",
			fields: fields{
				Name:   "undefine",
				Config: PreemptionConfig{Resource: []string{"undefine"}},
			},
			wantErr: true,
		},
		{
			name:    "TC3-empty config",
			wantErr: true,
		},
	}

	for _, tt := range validateTC {
		t.Run(tt.name, func(t *testing.T) {
			q := &Preemption{
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
				config: tt.fields.Config,
			}
			if err := q.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("QoS.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
