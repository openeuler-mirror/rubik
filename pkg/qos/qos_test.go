// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2021-04-17
// Description: QoS testing

package qos

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type getQosTestArgs struct {
	root string
	file string
}

type getQosTestCase struct {
	name    string
	args    getQosTestArgs
	want    int
	wantErr bool
}

const (
	qosFileWithValueNegativeOne string = "qos_level_with_negative_one"
	qosFileWithValueZero        string = "qos_level_with_value_zero"
	qosFileWithValueInvalid     string = "qos_level_with_value_invalid"
)

func newGetTestCases(qosDir string) []getQosTestCase {
	return []getQosTestCase{
		{
			name:    "TC1-get qos diff with value -1",
			args:    getQosTestArgs{root: qosDir, file: qosFileWithValueNegativeOne},
			want:    1,
			wantErr: true,
		},
		{
			name:    "TC2-get qos ok with value 0",
			args:    getQosTestArgs{root: qosDir, file: qosFileWithValueZero},
			want:    0,
			wantErr: false,
		},
		{
			name:    "TC3-get qos failed with invalid value",
			args:    getQosTestArgs{root: qosDir, file: qosFileWithValueInvalid},
			want:    1,
			wantErr: true,
		},
		{
			name:    "TC4-get qos failed with invalid file",
			args:    getQosTestArgs{root: qosDir, file: "file/not/exist"},
			want:    1,
			wantErr: true,
		},
		{
			name:    "TC5-get qos failed with not exist file",
			args:    getQosTestArgs{root: "/path/not/exist/file", file: "file_not_exist"},
			want:    1,
			wantErr: true,
		},
	}
}

// test_rubik_check_cgroup_qoslevel_with_podinfo_0001
func Test_getQos(t *testing.T) {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer os.RemoveAll(constant.TmpTestDir)
	qosDir, err := ioutil.TempDir(constant.TmpTestDir, "qos")
	assert.NoError(t, err)

	os.MkdirAll(filepath.Join(qosDir, "diff"), constant.DefaultDirMode)
	err = ioutil.WriteFile(filepath.Join(qosDir, qosFileWithValueNegativeOne), []byte("-1"), constant.DefaultFileMode)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(qosDir, "diff", qosFileWithValueNegativeOne), []byte("0"),
		constant.DefaultFileMode)
	assert.NoError(t, err)
	for _, dir := range []string{"", "diff"} {
		err = ioutil.WriteFile(filepath.Join(qosDir, dir, qosFileWithValueZero), []byte("0"),
			constant.DefaultFileMode)
		assert.NoError(t, err)
		err = ioutil.WriteFile(filepath.Join(qosDir, dir, qosFileWithValueInvalid), []byte("a"),
			constant.DefaultFileMode)
		assert.NoError(t, err)
	}

	tests := newGetTestCases(qosDir)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getQosLevel(tt.args.root, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("getQosLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getQosLevel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// test_rubik_set_cgroup_qoslevel_0001
// test_rubik_set_cgroup_qoslevel_0002
// test_rubik_set_cgroup_qoslevel_0003
func Test_validateQos(t *testing.T) {
	type args struct {
		qosLevel int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "TC1-valid qos level with value -1",
			args:    args{qosLevel: -1},
			wantErr: false,
		},
		{
			name:    "TC2-valid qos level with value 0",
			args:    args{qosLevel: 0},
			wantErr: false,
		},
		{
			name:    "TC3-invalid qos level",
			args:    args{qosLevel: 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkQosLevel(tt.args.qosLevel); (err != nil) != tt.wantErr {
				t.Errorf("checkQosLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type setQoSTestArgs struct {
	root     string
	file     string
	qosLevel int
}

type setQosTestCase struct {
	name    string
	args    setQoSTestArgs
	wantErr bool
}

func newSetTestCases(qosDir string, qosFilePath *os.File) []setQosTestCase {
	return []setQosTestCase{
		{
			name: "TC1-set qos ok with value -1",
			args: setQoSTestArgs{
				root:     qosDir,
				file:     "cpu",
				qosLevel: -1,
			},
			wantErr: false,
		},
		{
			name: "TC1.1-set qos not ok with previous value is -1",
			args: setQoSTestArgs{
				root:     qosDir,
				file:     "cpu",
				qosLevel: 0,
			},
			wantErr: true,
		},
		{
			name: "TC2-set qos not ok with empty cgroup path",
			args: setQoSTestArgs{
				root:     "",
				file:     "cpu",
				qosLevel: 0,
			},
			wantErr: true,
		},
		{
			name: "TC3-set qos not ok with invalid cgroup path",
			args: setQoSTestArgs{
				root:     qosFilePath.Name(),
				file:     "cpu",
				qosLevel: 0,
			},
			wantErr: true,
		},
	}
}

// Test_setQos is setQosLevel function test
func Test_setQos(t *testing.T) {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer os.RemoveAll(constant.TmpTestDir)
	qosDir, err := ioutil.TempDir(constant.TmpTestDir, "qos")
	assert.NoError(t, err)
	qosFilePath, err := ioutil.TempFile(qosDir, "qos_file")
	assert.NoError(t, err)

	tests := newSetTestCases(qosDir, qosFilePath)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setQosLevel(tt.args.root, tt.args.file, tt.args.qosLevel); (err != nil) != tt.wantErr {
				t.Errorf("setQosLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	err = qosFilePath.Close()
	assert.NoError(t, err)
}

type fields struct {
	CgroupRoot string
	CgroupPath string
}

type podInfoTestcase struct {
	name    string
	fields  fields
	want    map[string]string
	wantErr bool
}

func newPodInfoTestcases(cgRoot string) []podInfoTestcase {
	return []podInfoTestcase{
		{
			name: "TC1-get cgroup path ok with pre-define cgroupRoot",
			fields: fields{
				CgroupRoot: cgRoot,
				CgroupPath: "kubepods/podaaa",
			},
			want: map[string]string{"cpu": filepath.Join(cgRoot, "cpu", "kubepods/podaaa"),
				"memory": filepath.Join(cgRoot, "memory", "kubepods/podaaa")},
		},
		{
			name: "TC2-get cgroup path ok with non define cgroupRoot",
			fields: fields{
				CgroupPath: "kubepods/podbbb",
			},
			want: map[string]string{"cpu": filepath.Join(constant.DefaultCgroupRoot, "cpu",
				"kubepods/podbbb"), "memory": filepath.Join(constant.DefaultCgroupRoot, "memory", "kubepods/podbbb")},
		},
		{
			name:    "TC3-get invalid cgroup path",
			fields:  fields{CgroupPath: "invalid/cgroup/prefix/podbbb"},
			wantErr: true,
		},
		{
			name: "TC4-cgroup path too long",
			fields: fields{
				CgroupPath: "kubepods/cgroup/prefix/podbbb" + strings.Repeat("/long", constant.MaxCgroupPathLen),
			},
			wantErr: true,
		},
		{
			name:    "TC5-cgroup invalid cgroup path kubepods",
			fields:  fields{CgroupPath: "kubepods"},
			wantErr: true,
		},
		{
			name:    "TC6-cgroup invalid cgroup path kubepods/besteffort",
			fields:  fields{CgroupRoot: "", CgroupPath: "kubepods/besteffort/../besteffort"},
			wantErr: true,
		},
		{
			name:    "TC7-cgroup invalid cgroup path kubepods/burstable",
			fields:  fields{CgroupRoot: "", CgroupPath: "kubepods/burstable//"},
			wantErr: true,
		},
	}
}

// test_rubik_check_podinfo_0002
func TestPodInfo_CgroupFullPath(t *testing.T) {
	cgRoot := filepath.Join(constant.TmpTestDir, t.Name())

	tests := newPodInfoTestcases(cgRoot)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &PodInfo{
				CgroupRoot: tt.fields.CgroupRoot,
				PodQoS: api.PodQoS{
					CgroupPath: tt.fields.CgroupPath,
				},
			}
			err := pod.initCgroupPath()
			fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("initCgroupPath() = %v, want %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(pod.FullPath, tt.want) {
				t.Errorf("initCgroupPath() = %v, want %v", pod.FullPath, tt.want)
			}
		})
	}
}

type setQosFields struct {
	CgroupRoot string
	CgroupPath string
	QoSLevel   int
	PodID      string
}

type setQosTestCase2 struct {
	name            string
	fields          setQosFields
	wantSetErr      bool
	wantValidateErr bool
}

func newSetQosTestCase2(cgRoot string) []setQosTestCase2 {
	invalidQosLevel, repeatName := 999, 70
	return []setQosTestCase2{
		{
			name: "TC1-setup qos ok",
			fields: setQosFields{
				CgroupRoot: cgRoot,
				CgroupPath: "kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b2",
				QoSLevel:   -1, PodID: "poda5cb0d50-1234-1234-1234-e0ae4b7884b2",
			},
		},
		{
			name: "TC2-setup invalid qos value",
			fields: setQosFields{
				CgroupRoot: cgRoot,
				CgroupPath: "kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b3",
				QoSLevel:   invalidQosLevel, PodID: "poda5cb0d50-1234-1234-1234-e0ae4b7884b2",
			},
			wantValidateErr: true, wantSetErr: true,
		},
		{
			name: "TC3-setup too long podID",
			fields: setQosFields{
				CgroupRoot: cgRoot,
				CgroupPath: "kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b3",
				QoSLevel:   0, PodID: "poda5cb0d50" + strings.Repeat("-1234", repeatName),
			},
			wantValidateErr: true, wantSetErr: true,
		},
		{
			name: "TC4-setup invalid cgroupPath",
			fields: setQosFields{
				CgroupRoot: cgRoot,
				CgroupPath: "besteffort/poda5cb0d50-1234-1234-e0ae4b7884b2",
				QoSLevel:   -1, PodID: "poda5cb0d50-1234-1234-e0ae4b7884b2",
			},
			wantValidateErr: true, wantSetErr: true,
		},
		{
			name: "TC5-not exist qos file",
			fields: setQosFields{
				CgroupRoot: cgRoot,
				CgroupPath: "kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b3",
				QoSLevel:   -1, PodID: "poda5cb0d50-1234-1234-1234-e0ae4b7884b2",
			},
			wantValidateErr: true, wantSetErr: true,
		},
	}
}

// test_rubik_check_podinfo_0001
// test_rubik_check_cgroup_qoslevel_with_podinfo_0001
// test_rubik_check_cgroup_qoslevel_with_podinfo_0002
func TestPodInfo_SetQos(t *testing.T) {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	defer os.RemoveAll(constant.TmpTestDir)
	cgRoot, err := ioutil.TempDir(constant.TmpTestDir, t.Name())
	assert.NoError(t, err)
	podCPUCgroup := filepath.Join(cgRoot, "/cpu/kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b2")
	podMemoryCgroup := filepath.Join(cgRoot, "/memory/kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b2")
	tests := newSetQosTestCase2(cgRoot)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := api.PodQoS{CgroupPath: tt.fields.CgroupPath, QosLevel: tt.fields.QoSLevel}
			pod, err := NewPodInfo(context.Background(), tt.fields.PodID, tt.fields.CgroupRoot, req)
			if err != nil {
				if !tt.wantSetErr {
					t.Errorf("new PodInfo for %s failed: %v", tt.fields.PodID, err)
				}
				return
			}
			os.MkdirAll(podCPUCgroup, constant.DefaultFileMode)
			os.MkdirAll(podMemoryCgroup, constant.DefaultFileMode)
			if err := pod.SetQos(); (err != nil) != tt.wantSetErr {
				t.Errorf("SetQos() error = %v, wantErr %v", err, tt.wantSetErr)
			}
			if err := pod.ValidateQos(); (err != nil) != tt.wantValidateErr {
				t.Errorf("ValidateQos() error = %v, wantErr %v", err, tt.wantValidateErr)
			}
			err = os.RemoveAll(podCPUCgroup)
			assert.NoError(t, err)
			err = os.RemoveAll(podMemoryCgroup)
			assert.NoError(t, err)
		})
	}
	// test fullPath nil
	pod := &PodInfo{PodQoS: api.PodQoS{QosLevel: 0}, PodID: "abc", Ctx: context.Background()}
	err = pod.SetQos()
	assert.Equal(t, true, err != nil)
	// test cgroup qoslevel differ with pod qoslevel
	req := api.PodQoS{CgroupPath: "kubepods/besteffort/poda5cb0d50-1234-1234-1234-e0ae4b7884b2", QosLevel: 0}
	pod, err = NewPodInfo(context.Background(), "poda5cb0d50-1234-1234-1234-e0ae4b7884b2", cgRoot, req)
	assert.NoError(t, err)
	os.MkdirAll(podCPUCgroup, constant.DefaultFileMode)
	os.MkdirAll(podMemoryCgroup, constant.DefaultFileMode)
	err = pod.SetQos()
	assert.NoError(t, err)
	pod.QosLevel = -1
	err = pod.ValidateQos()
	assert.Equal(t, true, err != nil)
}

func TestBuildOfflinePodInfo(t *testing.T) {
	pod := corev1.Pod{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			UID: "podabc",
		},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			QOSClass: corev1.PodQOSGuaranteed,
		},
	}

	podQosInfo, err := BuildOfflinePodInfo(&pod)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/podpodabc")

	pod.Status.QOSClass = corev1.PodQOSBurstable
	podQosInfo, err = BuildOfflinePodInfo(&pod)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/burstable/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/burstable/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/burstable/podpodabc")

	pod.Status.QOSClass = corev1.PodQOSBestEffort
	podQosInfo, err = BuildOfflinePodInfo(&pod)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/besteffort/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/besteffort/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/besteffort/podpodabc")
}
