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
// Create: 2022-04-27
// Description: provide pods checkpoint management

// Package checkpoint provide pods checkpoint management.
package checkpoint

import (
	"path/filepath"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/config"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
	"isula.org/rubik/pkg/util"
)

// Checkpoint stores the binding between the CPU and pod container.
type Checkpoint struct {
	Pods map[string]*typedef.PodInfo `json:"pods,omitempty"`
}

// Manager manage checkpoint
type Manager struct {
	Checkpoint *Checkpoint
	sync.Mutex
}

// NewManager create manager
func NewManager() *Manager {
	return &Manager{
		Checkpoint: &Checkpoint{
			Pods: make(map[string]*typedef.PodInfo, 0),
		},
	}
}

// AddPod returns pod info from pod ID
func (cm *Manager) AddPod(pod *corev1.Pod) {
	// Before adding pod to the checkpoint, ensure that the pod is in the running state. Otherwise, problems may occur.
	// Only pod.Status.Phase == corev1.PodRunning
	cm.Lock()
	defer cm.Unlock()
	if pod == nil || string(pod.UID) == "" {
		return
	}
	if _, ok := cm.Checkpoint.Pods[string(pod.UID)]; ok {
		log.Debugf("pod %v is existed", string(pod.UID))
		return
	}
	log.Debugf("add pod %v", string(pod.UID))
	cm.Checkpoint.Pods[string(pod.UID)] = NewPodInfo(pod)
}

// GetPod returns pod info from pod ID
func (cm *Manager) GetPod(podID types.UID) *typedef.PodInfo {
	cm.Lock()
	defer cm.Unlock()
	return cm.Checkpoint.Pods[string(podID)]
}

// PodExist returns true if there is a pod whose key is podID in the checkpoint
func (cm *Manager) PodExist(podID types.UID) bool {
	cm.Lock()
	defer cm.Unlock()
	_, ok := cm.Checkpoint.Pods[string(podID)]
	return ok
}

// DelPod delete pod from checkpoint
func (cm *Manager) DelPod(podID types.UID) {
	cm.Lock()
	defer cm.Unlock()
	if _, ok := cm.Checkpoint.Pods[string(podID)]; !ok {
		log.Debugf("pod %v is not existed", string(podID))
		return
	}
	log.Debugf("delete pod %v", podID)
	delete(cm.Checkpoint.Pods, string(podID))
}

// UpdatePod updates pod information based on pods
func (cm *Manager) UpdatePod(pod *corev1.Pod) {
	cm.Lock()
	defer cm.Unlock()
	old, ok := cm.Checkpoint.Pods[string(pod.UID)]
	if !ok {
		log.Debugf("pod %v is not existed", string(pod.UID))
		return
	}
	log.Debugf("update pod %v", string(pod.UID))
	updatePodInfoNoLock(old, pod)
}

// SyncFromCluster synchronizing data from the kubernetes cluster using the list mechanism at the beginning
func (cm *Manager) SyncFromCluster(items []corev1.Pod) {
	cm.Lock()
	defer cm.Unlock()
	for _, pod := range items {
		if string(pod.UID) == "" {
			continue
		}
		log.Debugf("add pod %v", string(pod.UID))
		cm.Checkpoint.Pods[string(pod.UID)] = NewPodInfo(&pod)
	}
}

// filter filtering for list functions
type filter func(pi *typedef.PodInfo) bool

// listContainersWithFilters filters and returns deep copy objects of all containers
func (cm *Manager) listContainersWithFilters(filters ...filter) map[string]*typedef.ContainerInfo {
	cm.Lock()
	defer cm.Unlock()
	cc := make(map[string]*typedef.ContainerInfo, len(cm.Checkpoint.Pods))

	for _, pod := range cm.Checkpoint.Pods {
		if !mergeFilters(pod, filters) {
			continue
		}
		for _, ci := range pod.Containers {
			cc[ci.ID] = ci.Clone()
		}
	}

	return cc
}

func mergeFilters(pi *typedef.PodInfo, filters []filter) bool {
	for _, f := range filters {
		if !f(pi) {
			return false
		}
	}
	return true
}

// ListOfflineContainers filtering offline containers
func (cm *Manager) ListOfflineContainers() map[string]*typedef.ContainerInfo {
	return cm.listContainersWithFilters(func(pi *typedef.PodInfo) bool {
		return pi.Offline
	})
}

// ListAllContainers returns all containers copies
func (cm *Manager) ListAllContainers() map[string]*typedef.ContainerInfo {
	return cm.listContainersWithFilters()
}

// NewPodInfo create PodInfo
func NewPodInfo(pod *corev1.Pod) *typedef.PodInfo {
	pi := &typedef.PodInfo{
		Name:       pod.Name,
		UID:        string(pod.UID),
		Containers: make(map[string]*typedef.ContainerInfo, 0),
		CgroupPath: util.GetPodCgroupPath(pod),
	}
	updatePodInfoNoLock(pi, pod)
	return pi
}

// updatePodInfoNoLock updates PodInfo from the pod of Kubernetes.
// UpdatePodInfoNoLock does not lock pods during the modification.
// Therefore, ensure that the pod is being used only by this function.
// Currently, the checkpoint manager variable is locked when this function is invoked.
func updatePodInfoNoLock(pi *typedef.PodInfo, pod *corev1.Pod) {
	const (
		dockerPrefix     = "docker://"
		containerdPrefix = "containerd://"
	)
	pi.Name = pod.Name
	pi.Offline = util.IsOffline(pod)

	nameID := make(map[string]string, len(pod.Status.ContainerStatuses))
	for _, c := range pod.Status.ContainerStatuses {
		// Rubik is compatible with dockerd and containerd container engines.
		cid := strings.TrimPrefix(c.ContainerID, dockerPrefix)
		cid = strings.TrimPrefix(cid, containerdPrefix)

		// the container may be in the creation or deletion phase.
		if len(cid) == 0 {
			log.Debugf("no container id found of container %v", c.Name)
			continue
		}
		nameID[c.Name] = cid
	}
	// update ContainerInfo in a PodInfo
	for _, c := range pod.Spec.Containers {
		ci, ok := pi.Containers[c.Name]
		// add a container
		if !ok {
			log.Debugf("add new container %v", c.Name)
			pi.AddContainerInfo(typedef.NewContainerInfo(c, string(pod.UID), nameID[c.Name],
				config.CgroupRoot, pi.CgroupPath))
			continue
		}
		// The container name remains unchanged, and other information about the container is updated.
		ci.ID = nameID[c.Name]
		ci.CgroupAddr = filepath.Join(pi.CgroupPath, ci.ID)
	}
	// delete a container that does not exist
	for name := range pi.Containers {
		if _, ok := nameID[name]; !ok {
			log.Debugf("delete container %v", name)
			delete(pi.Containers, name)
		}
	}
}
