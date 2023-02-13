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
package qos

import (
	"context"
	"testing"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/test/try"
)

func init() {
	cgroup.InitMountDir(try.TestRoot)
}

type fields struct {
	Name   string
	Log    api.Logger
	Config Config
}
type args struct {
	old *typedef.PodInfo
	new *typedef.PodInfo
}

type test struct {
	name    string
	fields  fields
	args    args
	wantErr bool
	preHook func(*typedef.PodInfo) *typedef.PodInfo
}

var getCommonField = func(subSys []string) fields {
	return fields{
		Name:   "qos",
		Log:    log.WithCtx(context.WithValue(context.Background(), log.CtxKey(constant.LogEntryKey), "qos")),
		Config: Config{SubSys: subSys},
	}
}
var addFuncTC = []test{
	{
		name:   "TC1-set offline pod qos ok",
		fields: getCommonField([]string{"cpu", "memory"}),
		args: args{
			new: try.GenFakeOfflinePod([]*cgroup.Key{
				supportCgroupTypes["cpu"],
				supportCgroupTypes["memory"],
			}).PodInfo,
		},
	},
	{
		name:   "TC2-set online pod qos ok",
		fields: getCommonField([]string{"cpu", "memory"}),
		args: args{
			new: try.GenFakeOnlinePod([]*cgroup.Key{
				supportCgroupTypes["cpu"],
				supportCgroupTypes["memory"],
			}).WithContainers(3).PodInfo,
		},
	},
	{
		name:    "TC3-empty pod info",
		fields:  getCommonField([]string{"cpu", "memory"}),
		wantErr: true,
	},
	{
		name:   "TC4-invalid annotation key",
		fields: getCommonField([]string{"cpu"}),
		args: args{
			new: try.GenFakeBestEffortPod([]*cgroup.Key{supportCgroupTypes["cpu"]}).PodInfo,
		},
		preHook: func(pod *typedef.PodInfo) *typedef.PodInfo {
			newPod := pod.DeepCopy()
			newPod.Annotations["undefine"] = "true"
			return newPod
		},
	},
	{
		name:   "TC5-invalid annotation value",
		fields: getCommonField([]string{"cpu"}),
		args: args{
			new: try.GenFakeBestEffortPod([]*cgroup.Key{supportCgroupTypes["cpu"]}).PodInfo,
		},
		preHook: func(pod *typedef.PodInfo) *typedef.PodInfo {
			newPod := pod.DeepCopy()
			newPod.Annotations[constant.PriorityAnnotationKey] = "undefine"
			return newPod
		},
	},
}

func TestQoS_AddFunc(t *testing.T) {
	for _, tt := range addFuncTC {
		t.Run(tt.name, func(t *testing.T) {
			q := &QoS{
				Name:   tt.fields.Name,
				Log:    tt.fields.Log,
				Config: tt.fields.Config,
			}
			if tt.preHook != nil {
				tt.preHook(tt.args.new)
			}
			if err := q.AddFunc(tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("QoS.AddFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var updateFuncTC = []test{
	{
		name:   "TC1-online to offline",
		fields: getCommonField([]string{"cpu"}),
		args:   args{old: try.GenFakeOnlinePod([]*cgroup.Key{supportCgroupTypes["cpu"]}).WithContainers(3).PodInfo},
		preHook: func(pod *typedef.PodInfo) *typedef.PodInfo {
			newPod := pod.DeepCopy()
			// TODO: need fix pod.DeepCopy
			newAnnotation := make(map[string]string, 0)
			newAnnotation[constant.PriorityAnnotationKey] = "true"
			newPod.Annotations = newAnnotation
			return newPod
		},
	},
	{
		name:    "TC2-offline to online",
		fields:  getCommonField([]string{"cpu"}),
		args:    args{old: try.GenFakeOfflinePod([]*cgroup.Key{supportCgroupTypes["cpu"]}).PodInfo},
		wantErr: true,
	},
	{
		name:   "TC3-online to online",
		fields: getCommonField([]string{"cpu"}),
		args:   args{old: try.GenFakeOnlinePod([]*cgroup.Key{supportCgroupTypes["cpu"]}).PodInfo},
	},
}

func TestQoS_UpdateFunc(t *testing.T) {
	for _, tt := range updateFuncTC {
		t.Run(tt.name, func(t *testing.T) {
			q := &QoS{
				Name:   tt.fields.Name,
				Log:    tt.fields.Log,
				Config: tt.fields.Config,
			}
			if tt.preHook != nil {
				tt.args.new = tt.preHook(tt.args.old)
			}
			if err := q.UpdateFunc(tt.args.old, tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("QoS.UpdateFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var validateTC = []test{
	{
		name: "TC1-normal config",
		fields: fields{
			Name:   "qos",
			Log:    log.WithCtx(context.WithValue(context.Background(), log.CtxKey(constant.LogEntryKey), "qos")),
			Config: Config{SubSys: []string{"cpu", "memory"}},
		},
	},
	{
		name: "TC2-abnormal config",
		fields: fields{
			Name:   "undefine",
			Log:    log.WithCtx(context.WithValue(context.Background(), log.CtxKey(constant.LogEntryKey), "qos")),
			Config: Config{SubSys: []string{"undefine"}},
		},
		wantErr: true,
	},
	{
		name:    "TC3-empty config",
		wantErr: true,
	},
}

func TestQoS_Validate(t *testing.T) {
	for _, tt := range validateTC {
		t.Run(tt.name, func(t *testing.T) {
			q := &QoS{
				Name:   tt.fields.Name,
				Log:    tt.fields.Log,
				Config: tt.fields.Config,
			}
			if err := q.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("QoS.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
