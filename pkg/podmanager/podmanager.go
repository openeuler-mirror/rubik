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
// Create: 2023-01-12
// Description: This file defines PodManager passing and processing raw pod data

// Package podmanager implements manager connecting informer and module manager
package podmanager

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/subscriber"
	"isula.org/rubik/pkg/core/typedef"
)

// PodManagerName is the unique identity of PodManager
const PodManagerName = "DefaultPodManager"

// PodManager manages pod cache and pushes cache change events based on external input
type PodManager struct {
	api.Subscriber
	api.Publisher
	pods *podCache
}

// NewPodManager returns a PodManager pointer
func NewPodManager(publisher api.Publisher) *PodManager {
	manager := &PodManager{
		pods:      NewPodCache(),
		Publisher: publisher,
	}
	manager.Subscriber = subscriber.NewGenericSubscriber(manager, PodManagerName)
	return manager
}

// HandleEvent handles the event from publisher
func (manager *PodManager) HandleEvent(eventType typedef.EventType, event typedef.Event) {
	pod, err := eventToRawPod(event)
	if err != nil {
		log.Warnf(err.Error())
		return
	}
	switch eventType {
	case typedef.RAW_POD_ADD:
		manager.addFunc(pod)
	case typedef.RAW_POD_UPDATE:
		manager.updateFunc(pod)
	case typedef.RAW_POD_DELETE:
		manager.deleteFunc(pod)
	default:
		log.Infof("fail to process %s type event", eventType.String())
	}
}

// EventTypes returns the intersted event types
func (manager *PodManager) EventTypes() []typedef.EventType {
	return []typedef.EventType{typedef.RAW_POD_ADD, typedef.RAW_POD_UPDATE, typedef.RAW_POD_DELETE}
}

// eventToRawPod converts the event interface to RawPod pointer
func eventToRawPod(e typedef.Event) (*typedef.RawPod, error) {
	pod, ok := e.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("fail to get *typedef.RawPod which type is %T", e)
	}
	rawPod := typedef.RawPod(*pod)
	return &rawPod, nil
}

// addFunc handles the pod add event
func (manager *PodManager) addFunc(pod *typedef.RawPod) {
	// condition 1: only add running pod
	if !pod.Running() {
		log.Debugf("pod %v is not running", pod.UID)
		return
	}
	// condition2: pod is not existed
	if manager.pods.podExist(pod.ID()) {
		log.Debugf("pod %v has added", pod.UID)
		return
	}
	// step1: get pod information
	podInfo := pod.StripInfo()
	if podInfo == nil {
		log.Errorf("fail to strip info from raw pod")
		return
	}
	// step2. add pod information
	manager.tryAdd(podInfo)
}

// updateFunc handles the pod update event
func (manager *PodManager) updateFunc(pod *typedef.RawPod) {
	// step1: delete existed but not running pod
	if !pod.Running() {
		manager.tryDelete(pod.ID())
		return
	}

	// add or update information for running pod
	podInfo := pod.StripInfo()
	if podInfo == nil {
		log.Errorf("fail to strip info from raw pod")
		return
	}
	// The calling order must be updated first and then added
	// step2: process exsited and running pod
	manager.tryUpdate(podInfo)
	// step3: process not exsited and running pod
	manager.tryAdd(podInfo)
}

// deleteFunc handles the pod delete event
func (manager *PodManager) deleteFunc(pod *typedef.RawPod) {
	manager.tryDelete(pod.ID())
}

// tryAdd tries to add pod info which is not added
func (manager *PodManager) tryAdd(podInfo *typedef.PodInfo) {
	// only add when pod is not existed
	if !manager.pods.podExist(podInfo.UID) {
		manager.pods.addPod(podInfo)
		manager.Publish(typedef.INFO_ADD, podInfo)
	}
}

// tryUpdate tries to update podinfo which is existed
func (manager *PodManager) tryUpdate(podInfo *typedef.PodInfo) {
	// only update when pod is existed
	if manager.pods.podExist(podInfo.UID) {
		oldPod := manager.pods.getPod(podInfo.UID)
		manager.pods.updatePod(podInfo)
		manager.Publish(typedef.INFO_UPDATE, []*typedef.PodInfo{oldPod, podInfo.Clone()})
	}
}

// tryDelete tries to delete podinfo which is existed
func (manager *PodManager) tryDelete(id string) {
	// only delete when pod is existed
	oldPod := manager.pods.getPod(id)
	if oldPod != nil {
		manager.pods.delPod(id)
		manager.Publish(typedef.INFO_DELETE, oldPod)
	}
}
