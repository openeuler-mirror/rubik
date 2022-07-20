// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2022-05-10
// Description: checkpoint DT test

package checkpoint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/typedef"
)

var containerInfos = []*typedef.ContainerInfo{
	{
		Name:       "FooCon",
		ID:         "testCon1",
		PodID:      "testPod1",
		CgroupRoot: constant.DefaultCgroupRoot,
		CgroupAddr: "kubepods/testPod1/testCon1",
	},
	{
		Name:       "BarCon",
		ID:         "testCon2",
		PodID:      "testPod2",
		CgroupRoot: constant.DefaultCgroupRoot,
		CgroupAddr: "kubepods/testPod2/testCon2",
	},
	{
		Name:       "BiuCon",
		ID:         "testCon3",
		PodID:      "testPod3",
		CgroupRoot: constant.DefaultCgroupRoot,
		CgroupAddr: "kubepods/testPod3/testCon3",
	},
	{
		Name:       "PahCon",
		ID:         "testCon4",
		PodID:      "testPod4",
		CgroupRoot: constant.DefaultCgroupRoot,
		CgroupAddr: "kubepods/testPod4/testCon4",
	},
}

var podInfos = []*typedef.PodInfo{
	// allow quota adjustment
	{
		Name: "FooPod",
		UID:  containerInfos[0].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			containerInfos[0].Name: containerInfos[0],
		},
	},
	// allow quota adjustment
	{
		Name: "BarPod",
		UID:  containerInfos[1].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			containerInfos[1].Name: containerInfos[1],
		},
	},
	// quota adjustment is not allowed
	{
		Name: "BiuPod",
		UID:  containerInfos[2].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			containerInfos[2].Name: containerInfos[2],
		},
	},
	// quota adjustment is not allowed
	{
		Name: "PahPod",
		UID:  containerInfos[3].PodID,
		Containers: map[string]*typedef.ContainerInfo{
			containerInfos[3].Name: containerInfos[3],
		},
	},
}

var coreV1Pods = []corev1.Pod{
	{
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			UID:  types.UID("testPod5"),
			Name: "BiuPod",
		},
		Status: corev1.PodStatus{
			Phase:    corev1.PodRunning,
			QOSClass: corev1.PodQOSGuaranteed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "BiuCon",
					ContainerID: "docker://testCon5",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "BiuCon",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu": *resource.NewQuantity(2, resource.DecimalSI),
						},
						Limits: corev1.ResourceList{
							"cpu":    *resource.NewQuantity(3, resource.DecimalSI),
							"memory": *resource.NewQuantity(300, resource.DecimalSI),
						},
					},
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			UID:  types.UID(podInfos[1].UID),
			Name: podInfos[1].Name,
		},
		Status: corev1.PodStatus{
			Phase:    corev1.PodRunning,
			QOSClass: corev1.PodQOSGuaranteed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "BarCon1",
					ContainerID: "docker://testCon6",
				},
				{
					Name: "BarCon2",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "BarCon1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu": *resource.NewQuantity(2, resource.DecimalSI),
						},
						Limits: corev1.ResourceList{
							"cpu":    *resource.NewQuantity(2, resource.DecimalSI),
							"memory": *resource.NewQuantity(100, resource.DecimalSI),
						},
					},
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			UID:  types.UID(podInfos[0].UID),
			Name: "FooPod",
		},
		Status: corev1.PodStatus{
			Phase:    corev1.PodRunning,
			QOSClass: corev1.PodQOSGuaranteed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "FooCon",
					ContainerID: "docker://" + containerInfos[0].ID,
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "FooCon",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu": *resource.NewQuantity(2, resource.DecimalSI),
						},
						Limits: corev1.ResourceList{
							"cpu":    *resource.NewQuantity(3, resource.DecimalSI),
							"memory": *resource.NewQuantity(300, resource.DecimalSI),
						},
					},
				},
			},
		},
	},
}

// TestManagerAddPod tests AddPod of Manager
func TestManagerAddPod(t *testing.T) {
	var (
		podNum1 = 1
		podNum2 = 2
	)
	cpm := &Manager{
		Checkpoint: &Checkpoint{
			Pods: map[string]*typedef.PodInfo{
				podInfos[0].UID: podInfos[0].Clone(),
			},
		},
	}
	// 1. add pods that do not exist
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum1)
	var mockPahPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:  types.UID(podInfos[3].UID),
			Name: podInfos[3].Name,
		},
		Status: corev1.PodStatus{
			Phase:    corev1.PodRunning,
			QOSClass: corev1.PodQOSBurstable,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "PahCon",
					ContainerID: "docker://" + containerInfos[3].ID,
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "PahCon",
				},
			},
		},
	}
	cpm.AddPod(mockPahPod)
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum2)
	// 2.join a joined pods
	cpm.AddPod(mockPahPod)
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum2)

	// 3.add a pod whose name is empty
	cpm.AddPod(&coreV1Pods[0])
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum2)
}

// TestManagerDelPod tests DelPod of Manager
func TestManagerDelPod(t *testing.T) {
	var (
		podNum0 = 0
		podNum1 = 1
	)
	cpm := &Manager{
		Checkpoint: &Checkpoint{
			Pods: map[string]*typedef.PodInfo{
				podInfos[0].UID: podInfos[0].Clone(),
			},
		},
	}
	// 1. delete pods that do not exist
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum1)
	cpm.DelPod(types.UID(podInfos[1].UID))
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum1)
	// 2. delete existed pods
	cpm.DelPod(types.UID(podInfos[0].UID))
	assert.Equal(t, len(cpm.Checkpoint.Pods), podNum0)
}

// TestManagerUpdatePod test UpdatePod function of Manager
func TestManagerUpdatePod(t *testing.T) {
	var managerUpdatePodTests = []struct {
		pod       *corev1.Pod
		judgement func(t *testing.T, m *Manager)
		name      string
	}{
		{
			name: "TC1 - update a non-added pod",
			pod:  &coreV1Pods[1],
			judgement: func(t *testing.T, m *Manager) {
				podNum2 := 2
				assert.Equal(t, podNum2, len(m.Checkpoint.Pods))
			},
		},
	}
	cpm := &Manager{
		Checkpoint: &Checkpoint{
			Pods: map[string]*typedef.PodInfo{
				podInfos[0].UID: podInfos[0].Clone(),
				podInfos[1].UID: podInfos[1].Clone(),
			},
		},
	}

	for _, tt := range managerUpdatePodTests {
		t.Run(tt.name, func(t *testing.T) {
			cpm.UpdatePod(tt.pod)
			tt.judgement(t, cpm)
		})
	}
}

// TestManagerListPodsAndContainers tests methods of list pods and containers of Manager
func TestManagerListPodsAndContainers(t *testing.T) {
	var (
		podNum3 = 3
		podNum4 = 4
	)
	// The pod names in Kubernetes must be unique. The same pod cannot have the same name, but different pods can have the same name.
	cpm := &Manager{
		Checkpoint: &Checkpoint{
			Pods: map[string]*typedef.PodInfo{
				podInfos[0].UID: podInfos[0].Clone(),
				podInfos[1].UID: podInfos[1].Clone(),
				podInfos[2].UID: podInfos[2].Clone(),
			},
		},
	}
	// 1. Containers with Different Pods with Different Names
	assert.Equal(t, len(cpm.ListAllPods()), podNum3)
	assert.Equal(t, len(cpm.ListAllContainers()), podNum3)
	// 2. Containers with the same name in different pods
	var podWithSameNameCon = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:  types.UID("testPod5"),
			Name: "FakeFooPod",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "FooCon",
					ContainerID: "docker://testCon5",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "FooCon",
				},
			},
		},
	}

	cpm.AddPod(podWithSameNameCon)
	assert.Equal(t, len(cpm.ListAllContainers()), podNum4)
}

// TestManagerSyncFromCluster tests SyncFromCluster of Manager
func TestManagerSyncFromCluster(t *testing.T) {
	cpm := NewManager("")
	cpm.Checkpoint = &Checkpoint{
		Pods: make(map[string]*typedef.PodInfo, 0),
	}

	cpm.SyncFromCluster(coreV1Pods)
	expPodNum := 3
	assert.Equal(t, len(cpm.Checkpoint.Pods), expPodNum)

	pi2 := cpm.GetPod(coreV1Pods[1].UID)
	assert.Equal(t, "BiuPod", pi2.Name)
}

//  TestMangerPodExist tests the PodExist of Manger
func TestMangerPodExist(t *testing.T) {
	tests := []struct {
		name string
		id   types.UID
		want bool
	}{
		{
			name: "TC1 - check a non-existed pod",
			id:   types.UID(podInfos[0].UID),
			want: true,
		},
		{
			name: "TC2 - check an existed pod",
			id:   types.UID(podInfos[1].UID),
			want: false,
		},
	}
	cpm := NewManager("")
	cpm.Checkpoint = &Checkpoint{
		Pods: map[string]*typedef.PodInfo{
			podInfos[0].UID: podInfos[0].Clone(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cpm.PodExist(tt.id))
		})
	}
}
