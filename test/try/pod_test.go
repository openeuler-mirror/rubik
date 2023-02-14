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
// Description: This file contains pod info and cgroup construct

package try

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	corev1 "k8s.io/api/core/v1"
)

func TestNewFakePod(t *testing.T) {
	id := constant.PodCgroupNamePrefix + uuid.New().String()
	type args struct {
		keys     map[*cgroup.Key]string
		qosClass corev1.PodQOSClass
	}
	tests := []struct {
		name string
		args args
		want *FakePod
	}{
		{
			name: "TC1-new fake best effort pod",
			args: args{
				keys:     map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				qosClass: corev1.PodQOSBestEffort,
			},
			want: &FakePod{
				Keys: map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				PodInfo: &typedef.PodInfo{
					Name:       "fakepod-" + id[:idLen],
					UID:        id,
					Namespace:  "test",
					CgroupPath: filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)), id),
				},
			},
		},
		{
			name: "TC2-new fake guaranteed pod",
			args: args{
				keys:     map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				qosClass: corev1.PodQOSGuaranteed,
			},
			want: &FakePod{
				Keys: map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				PodInfo: &typedef.PodInfo{
					Name:       "fakepod-" + id[:idLen],
					UID:        id,
					Namespace:  "test",
					CgroupPath: filepath.Join(constant.KubepodsCgroup, id),
				},
			},
		},
		{
			name: "TC3-new fake burstable pod",
			args: args{
				keys:     map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				qosClass: corev1.PodQOSBurstable,
			},
			want: &FakePod{
				Keys: map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				PodInfo: &typedef.PodInfo{
					Name:       "fakepod-" + id[:idLen],
					UID:        id,
					Namespace:  "test",
					CgroupPath: filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)), id),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakePod := NewFakePod(tt.args.keys, tt.args.qosClass)
			assert.Equal(t, fakePod.Namespace, tt.want.Namespace)
			assert.Equal(t, len(fakePod.Name), len(tt.want.Name))
			assert.Equal(t, len(fakePod.UID), len(tt.want.UID))
			assert.Equal(t, len(fakePod.CgroupPath), len(tt.want.CgroupPath))
		})
	}
}

func TestGenFakePod(t *testing.T) {
	type args struct {
		keys         map[*cgroup.Key]string
		qosClass     corev1.PodQOSClass
		containerNum int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TC1-generate burstable pod",
			args: args{
				keys:     map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				qosClass: corev1.PodQOSBestEffort,
			},
		},
		{
			name: "TC2-generate guaranteed pod with 3 containers",
			args: args{
				keys:         map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"},
				qosClass:     corev1.PodQOSGuaranteed,
				containerNum: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakePod := GenFakePod(tt.args.keys, tt.args.qosClass)
			if tt.args.qosClass != corev1.PodQOSGuaranteed {
				// guaranteed pod does not have path prefix like "guaranteed/podxxx"
				assert.Equal(t, true, strings.Contains(fakePod.CgroupPath, strings.ToLower(string(corev1.PodQOSBestEffort))))

			}
			for key, val := range tt.args.keys {
				podCgroupFile := cgroup.AbsoluteCgroupPath(key.SubSys, fakePod.CgroupPath, key.FileName)
				assert.Equal(t, true, util.PathExist(podCgroupFile))
				ret := ReadFile(podCgroupFile)
				assert.NoError(t, ret.err)
				assert.Equal(t, val, ret.val)
			}
			if tt.args.containerNum != 0 {
				fakePod.WithContainers(tt.args.containerNum)
				for key, val := range tt.args.keys {
					for _, c := range fakePod.IDContainersMap {
						containerCgroupFile := cgroup.AbsoluteCgroupPath(key.SubSys, c.CgroupPath, key.FileName)
						assert.Equal(t, true, util.PathExist(containerCgroupFile))
						ret := ReadFile(containerCgroupFile)
						assert.NoError(t, ret.err)
						assert.Equal(t, val, ret.val)
					}
				}
			}
			fakePod.CleanPath().OrDie()
		})
	}
}

func TestGenParticularFakePod(t *testing.T) {
	type args struct {
		keys map[*cgroup.Key]string
	}
	tests := []struct {
		name           string
		kind           string
		wantAnnotation string
	}{
		{
			name:           "TC1-generate online pod",
			kind:           "online",
			wantAnnotation: "false",
		},
		{
			name:           "TC2-generate offline pod",
			kind:           "offline",
			wantAnnotation: "true",
		},
		{
			name:           "TC3-generate burstable pod",
			kind:           "burstable",
			wantAnnotation: "",
		},
		{
			name:           "TC4-generate besteffort pod",
			kind:           "besteffort",
			wantAnnotation: "",
		},
		{
			name:           "TC5-generate guaranteed pod",
			kind:           "guaranteed",
			wantAnnotation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"}
			var fakePod *FakePod
			switch tt.kind {
			case "online":
				fakePod = GenFakeOnlinePod(keys)
			case "offline":
				fakePod = GenFakeOfflinePod(keys)
			case "burstable":
				fakePod = GenFakeBurstablePod(keys)
			case "besteffort":
				fakePod = GenFakeBestEffortPod(keys)
			case "guaranteed":
				fakePod = GenFakeGuaranteedPod(keys)
			}
			assert.Equal(t, tt.wantAnnotation, fakePod.Annotations[constant.PriorityAnnotationKey])
			fakePod.CleanPath().OrDie()
		})
	}
}

func TestFakePod_DeepCopy(t *testing.T) {
	type fields struct {
		PodInfo *typedef.PodInfo
		Keys    map[*cgroup.Key]string
	}

	keys := map[*cgroup.Key]string{{SubSys: "cpu", FileName: constant.CPUCgroupFileName}: "0"}
	podInfo := NewFakePod(keys, corev1.PodQOSGuaranteed).PodInfo
	tests := []struct {
		name   string
		fields fields
		want   *FakePod
	}{
		{
			name: "TC1-deep copy",
			fields: fields{
				PodInfo: podInfo,
				Keys:    keys,
			},
			want: &FakePod{
				PodInfo: podInfo,
				Keys:    keys,
			},
		},
		{
			name: "TC2-empty copy",
			fields: fields{
				PodInfo: nil,
				Keys:    nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &FakePod{
				PodInfo: tt.fields.PodInfo,
				Keys:    tt.fields.Keys,
			}
			if got := pod.DeepCopy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FakePod.DeepCopy() = %v, want %v", got, tt.want)
			}
		})
	}
}
