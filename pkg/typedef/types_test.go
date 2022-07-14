// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jing Rui
// Create: 2022-07-10
// Description: This file contains default constants used in the project

// Package typedef is general used types.
package typedef

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func init() {
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	if err != nil {
		log.Fatalf("Failed to create tmp test dir for testing!")
	}
}

func genContainer() corev1.Container {
	c := corev1.Container{}
	c.Name = "testContainer"
	c.Resources.Requests = make(corev1.ResourceList)
	c.Resources.Limits = make(corev1.ResourceList)
	c.Resources.Requests["cpu"] = *resource.NewMilliQuantity(10000, resource.DecimalSI)
	c.Resources.Limits["cpu"] = *resource.NewMilliQuantity(10000, resource.DecimalSI)
	c.Resources.Limits["memory"] = *resource.NewMilliQuantity(10000, resource.DecimalSI)

	return c
}

// TestNewContainerInfo is testcase for NewContainerInfo
func TestNewContainerInfo(t *testing.T) {
	cgRoot, err := ioutil.TempDir(constant.TmpTestDir, "cgRoot")
	assert.NoError(t, err)
	defer os.RemoveAll(cgRoot)
	podCGPath, err := ioutil.TempDir(constant.TmpTestDir, "pod")
	assert.NoError(t, err)
	defer os.RemoveAll(cgRoot)

	c := genContainer()
	type args struct {
		container     corev1.Container
		podID         string
		conID         string
		cgroupRoot    string
		podCgroupPath string
	}
	tests := []struct {
		want *ContainerInfo
		name string
		args args
	}{
		{
			name: "TC",
			args: args{container: c, podID: "podID", cgroupRoot: cgRoot, conID: "cID", podCgroupPath: podCGPath},
			want: &ContainerInfo{
				Name:       "testContainer",
				ID:         "cID",
				PodID:      "podID",
				CgroupRoot: cgRoot,
				CgroupAddr: filepath.Join(podCGPath, "cID"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewContainerInfo(tt.args.container, tt.args.podID, tt.args.conID, tt.args.cgroupRoot, tt.args.podCgroupPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContainerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainerInfo_CgroupPath is testcase for ContainerInfo.CgroupPath
func TestContainerInfo_CgroupPath(t *testing.T) {
	cgRoot, err := ioutil.TempDir(constant.TmpTestDir, "cgRoot")
	assert.NoError(t, err)
	defer os.RemoveAll(cgRoot)
	podCGPath, err := ioutil.TempDir(constant.TmpTestDir, "pod")
	assert.NoError(t, err)
	defer os.RemoveAll(podCGPath)

	emptyCi := &ContainerInfo{}
	assert.Equal(t, "", emptyCi.CgroupPath("cpu"))

	ci := emptyCi.Clone()

	ci.Name = "testContainer"
	ci.ID = "cID"
	ci.PodID = "podID"
	ci.CgroupRoot = cgRoot
	ci.CgroupAddr = filepath.Join(podCGPath, "cID")
	assert.Equal(t, ci.CgroupPath("cpu"),
		filepath.Join(cgRoot, "cpu", filepath.Join(podCGPath, "cID")))
}

// TestPodInfo_Clone is testcase for PodInfo.Clone
func TestPodInfo_Clone(t *testing.T) {
	cgRoot, err := ioutil.TempDir(constant.TmpTestDir, "cgRoot")
	assert.NoError(t, err)
	defer os.RemoveAll(cgRoot)
	podCGPath, err := ioutil.TempDir(constant.TmpTestDir, "pod")
	assert.NoError(t, err)
	defer os.RemoveAll(podCGPath)
	emptyPI := &PodInfo{}
	pi := emptyPI.Clone()
	pi.Containers = make(map[string]*ContainerInfo)
	pi.Name = "testPod"
	pi.UID = "abcd"
	pi.CgroupPath = cgRoot

	containerWithOutName := genContainer()
	containerWithOutName.Name = ""

	emptyNameCI := NewContainerInfo(containerWithOutName, "testPod", "cID", cgRoot, podCGPath)
	pi.AddContainerInfo(emptyNameCI)
	assert.Equal(t, len(pi.Containers), 0)

	ci := NewContainerInfo(genContainer(), "testPod", "cID", cgRoot, podCGPath)
	pi.AddContainerInfo(ci)
	newPi := pi.Clone()
	assert.Equal(t, len(newPi.Containers), 1)
}
