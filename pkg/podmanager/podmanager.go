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

	nriapi "github.com/containerd/nri/pkg/api"
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
	Pods *PodCache
}

// NewPodManager returns a PodManager pointer
func NewPodManager(publisher api.Publisher) *PodManager {
	manager := &PodManager{
		Pods:      NewPodCache(),
		Publisher: publisher,
	}
	manager.Subscriber = subscriber.NewGenericSubscriber(manager, PodManagerName)
	return manager
}

// HandleEvent handles the event from publisher
func (manager *PodManager) HandleEvent(eventType typedef.EventType, event typedef.Event) {
	switch eventType {
	case typedef.RAWPODADD, typedef.RAWPODUPDATE, typedef.RAWPODDELETE:
		manager.handleWatchEvent(eventType, event)
	case typedef.RAWPODSYNCALL:
		manager.handleListEvent(eventType, event)
	case typedef.NRIPODADD, typedef.NRIPODDELETE:
		manager.handleNRIPodEvent(eventType, event)
	case typedef.NRICONTAINERSTART, typedef.NRICONTAINERREMOVE:
		manager.handleNRIContainerEvent(eventType, event)
	case typedef.NRIPODSYNCALL:
		manager.handleSYNCNRIPodsEvent(eventType, event)
	case typedef.NRICONTAINERSYNCALL:
		manager.handleSYNCNRIContainersEvent(eventType, event)
	default:
		log.Infof("failed to process %s type event", eventType.String())
	}
}

// handleWatchEvent handles the watch event
func (manager *PodManager) handleWatchEvent(eventType typedef.EventType, event typedef.Event) {
	pod, err := eventToRawPod(event)
	if err != nil {
		log.Warnf(err.Error())
		return
	}

	switch eventType {
	case typedef.RAWPODADD:
		manager.addFunc(pod)
	case typedef.RAWPODUPDATE:
		manager.updateFunc(pod)
	case typedef.RAWPODDELETE:
		manager.deleteFunc(pod)
	default:
		log.Errorf("invalid event type...")
	}
}

// handlenripodevent handles the nri pod event
func (manager *PodManager) handleNRIPodEvent(eventType typedef.EventType, event typedef.Event) {
	pod, err := eventToNRIRawPod(event)
	if err != nil {
		log.Warnf(err.Error())
		return
	}

	switch eventType {
	case typedef.NRIPODADD:
		manager.addNRIPodFunc(pod)
	case typedef.NRIPODDELETE:
		manager.deleteNRIPodFunc(pod)
	default:
		log.Errorf("code problem, should not go here...")
	}
}

// handlenricontainerevent handles the nri container event
func (manager *PodManager) handleNRIContainerEvent(eventType typedef.EventType, event typedef.Event) {
	container, err := eventToNRIRawContainer(event)
	if err != nil {
		log.Warnf(err.Error())
		return
	}
	switch eventType {
	case typedef.NRICONTAINERSTART:
		manager.addNRIContainerFunc(container)
	case typedef.NRICONTAINERREMOVE:
		manager.removeNRIContainerFunc(container)
	default:
		log.Errorf("code Problem, should not go here...")
	}

}

// handleListEvent handles the list event
func (manager *PodManager) handleListEvent(eventType typedef.EventType, event typedef.Event) {
	pods, err := eventToRawPods(event)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	switch eventType {
	case typedef.RAWPODSYNCALL:
		manager.sync(pods)
	default:
		log.Errorf("invalid event type...")
	}
}

// handleSYNCNRIPodsEvent handles sync pod event
func (manager *PodManager) handleSYNCNRIPodsEvent(eventType typedef.EventType, event typedef.Event) {
	pods, err := eventToNRIRawPods(event)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	switch eventType {
	case typedef.NRIPODSYNCALL:
		manager.nripodssync(pods)
	default:
		log.Errorf("code problem, should not go here...")
	}
}

// handleSYNCNRIContainersEvent handles sync container event
func (manager *PodManager) handleSYNCNRIContainersEvent(eventType typedef.EventType, event typedef.Event) {
	containers, err := eventToNRIRawContainers(event)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	switch eventType {
	case typedef.NRICONTAINERSYNCALL:
		manager.nricontainerssync(containers)
	default:
		log.Errorf("code problem, should not go here...")
	}
}

// eventToNRIRawContainers handles nri containers event
func eventToNRIRawContainers(e typedef.Event) ([]*typedef.NRIRawContainer, error) {
	containers, ok := e.([]*nriapi.Container)
	if !ok {
		return nil, fmt.Errorf("fail to get *typedef.NRIRawContainer which type is %T", e)
	}
	toRawContainerPointer := func(container nriapi.Container) *typedef.NRIRawContainer {
		tmp := typedef.NRIRawContainer(container)
		return &tmp
	}
	var pointerContainers []*typedef.NRIRawContainer
	for _, container := range containers {
		pointerContainers = append(pointerContainers, toRawContainerPointer(*container))
	}
	return pointerContainers, nil
}

// eventToNRIRawPods handles nri pods event
func eventToNRIRawPods(e typedef.Event) ([]*typedef.NRIRawPod, error) {
	pods, ok := e.([]*nriapi.PodSandbox)
	if !ok {
		return nil, fmt.Errorf("fail to get *typedef.NRIRawPod which type is %T", e)
	}
	toRawPodPointer := func(pod nriapi.PodSandbox) *typedef.NRIRawPod {
		tmp := typedef.NRIRawPod(pod)
		return &tmp
	}
	var pointerPods []*typedef.NRIRawPod
	for _, pod := range pods {
		pointerPods = append(pointerPods, toRawPodPointer(*pod))
	}
	return pointerPods, nil
}

// eventToNRIRawPod handles nri pod event
func eventToNRIRawPod(e typedef.Event) (*typedef.NRIRawPod, error) {
	pod, ok := e.(*nriapi.PodSandbox)
	if !ok {
		return nil, fmt.Errorf("fail to get *typedef.NRIRawPod which type is %T", e)
	}
	nriRawPod := typedef.NRIRawPod(*pod)
	return &nriRawPod, nil
}

// eventToNRIRawContainer handles nri container event
func eventToNRIRawContainer(e typedef.Event) (*typedef.NRIRawContainer, error) {
	container, ok := e.(*nriapi.Container)
	if !ok {
		return nil, fmt.Errorf("fail to get *typedef.NRIRawContainer which type is %T", e)
	}
	nriRawContainer := typedef.NRIRawContainer(*container)
	return &nriRawContainer, nil
}

// EventTypes returns the intersted event types
func (manager *PodManager) EventTypes() []typedef.EventType {
	return []typedef.EventType{typedef.RAWPODADD,
		typedef.RAWPODUPDATE,
		typedef.RAWPODDELETE,
		typedef.RAWPODSYNCALL,
		typedef.NRIPODADD,
		typedef.NRICONTAINERSTART,
		typedef.NRIPODDELETE,
		typedef.NRIPODSYNCALL,
		typedef.NRICONTAINERSYNCALL,
		typedef.NRICONTAINERREMOVE,
	}
}

// eventToRawPod converts the event interface to RawPod pointer
func eventToRawPod(e typedef.Event) (*typedef.RawPod, error) {
	pod, ok := e.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("failed to get raw pod information")
	}
	rawPod := typedef.RawPod(*pod)
	return &rawPod, nil
}

// eventToRawPods converts the event interface to RawPod pointer slice
func eventToRawPods(e typedef.Event) ([]*typedef.RawPod, error) {
	pods, ok := e.([]corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("failed to get raw pod information")
	}
	toRawPodPointer := func(pod corev1.Pod) *typedef.RawPod {
		tmp := typedef.RawPod(pod)
		return &tmp
	}
	var pointerPods []*typedef.RawPod
	for _, pod := range pods {
		pointerPods = append(pointerPods, toRawPodPointer(pod))
	}
	return pointerPods, nil
}

// addNRIPodFunc handles nri pod add event
func (manager *PodManager) addNRIPodFunc(pod *typedef.NRIRawPod) {
	// condition 1: only add running pod
	if !pod.Running() {
		log.Debugf("pod %v is not running", pod.Uid)
		return
	}
	// condition2: pod is not existed
	if manager.Pods.podExist(pod.ID()) {
		log.Debugf("pod %v has added", pod.Uid)
		return
	}
	// step1: get pod information
	podInfo := pod.ConvertNRIRawPod2PodInfo()
	if podInfo == nil {
		log.Errorf("fail to strip info from raw pod")
		return
	}
	// step2. add pod information
	manager.tryAddNRIPod(podInfo)
}

// addNRIContainerFunc handles add nri container event
func (manager *PodManager) addNRIContainerFunc(container *typedef.NRIRawContainer) {
	containerInfo := container.ConvertNRIRawContainer2ContainerInfo()
	for _, pod := range manager.Pods.Pods {
		if containerInfo.PodSandboxId == pod.ID {
			pod.IDContainersMap[containerInfo.ID] = containerInfo
			manager.Publish(typedef.INFOADD, pod.DeepCopy())
		}
	}
}

// sync to podCache after remove container
func (manager *PodManager) removeNRIContainerFunc(container *typedef.NRIRawContainer) {
	containerInfo := container.ConvertNRIRawContainer2ContainerInfo()
	for _, pod := range manager.Pods.Pods {
		if containerInfo.PodSandboxId == pod.ID {
			delete(pod.IDContainersMap, containerInfo.ID)
		}
	}
}

// addFunc handles the pod add event
func (manager *PodManager) addFunc(pod *typedef.RawPod) {
	// condition 1: only add running pod
	if !pod.Running() {
		log.Debugf("pod %v is not running", pod.UID)
		return
	}
	// condition2: pod is not existed
	if manager.Pods.podExist(pod.ID()) {
		log.Debugf("pod %v has already added", pod.UID)
		return
	}
	// step1: get pod information
	podInfo := pod.ExtractPodInfo()
	if podInfo == nil {
		log.Errorf("failed to extract information from raw pod")
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
	podInfo := pod.ExtractPodInfo()
	if podInfo == nil {
		log.Errorf("failed to extract information from raw pod")
		return
	}
	// The calling order must be updated first and then added
	// step2: process exited and running pod
	manager.tryUpdate(podInfo)
	// step3: process not exited and running pod
	manager.tryAdd(podInfo)
}

// deleteNRIPodFunc handles delete nri pod
func (manager *PodManager) deleteNRIPodFunc(pod *typedef.NRIRawPod) {
	manager.tryDelete(pod.ID())
}

// deleteFunc handles the pod delete event
func (manager *PodManager) deleteFunc(pod *typedef.RawPod) {
	manager.tryDelete(pod.ID())
}

// tryAdd tries to add pod info which is not added
func (manager *PodManager) tryAdd(podInfo *typedef.PodInfo) {
	// only add when pod is not existed
	if !manager.Pods.podExist(podInfo.UID) {
		manager.Pods.addPod(podInfo)
		manager.Publish(typedef.INFOADD, podInfo.DeepCopy())
	}
}

// tryAddNRIPod tries to add nri pod info which is not added
func (manager *PodManager) tryAddNRIPod(podInfo *typedef.PodInfo) {
	// only add when pod is not existed
	if !manager.Pods.podExist(podInfo.UID) {
		manager.Pods.addPod(podInfo)
	}
}

// tryUpdate tries to update podinfo which is existed
func (manager *PodManager) tryUpdate(podInfo *typedef.PodInfo) {
	// only update when pod is existed
	if manager.Pods.podExist(podInfo.UID) {
		oldPod := manager.Pods.getPod(podInfo.UID)
		manager.Pods.updatePod(podInfo)
		manager.Publish(typedef.INFOUPDATE, []*typedef.PodInfo{oldPod, podInfo.DeepCopy()})
	}
}

// tryDelete tries to delete podinfo which is existed
func (manager *PodManager) tryDelete(id string) {
	// only delete when pod is existed
	oldPod := manager.Pods.getPod(id)
	if oldPod != nil {
		manager.Pods.delPod(id)
		manager.Publish(typedef.INFODELETE, oldPod)
	}
}

// sync replaces all Pod information sent over
func (manager *PodManager) sync(pods []*typedef.RawPod) {
	var newPods []*typedef.PodInfo
	for _, pod := range pods {
		if pod == nil || !pod.Running() {
			continue
		}
		newPods = append(newPods, pod.ExtractPodInfo())
	}
	manager.Pods.substitute(newPods)
}

// nripodssync handles sync all pods
func (manager *PodManager) nripodssync(pods []*typedef.NRIRawPod) {
	var newPods []*typedef.PodInfo
	for _, pod := range pods {
		if pod == nil || !pod.Running() {
			continue
		}
		newPods = append(newPods, pod.ConvertNRIRawPod2PodInfo())
	}
	manager.Pods.substitute(newPods)
}

// nricontainerssync handles sync all containers
func (manager *PodManager) nricontainerssync(containers []*typedef.NRIRawContainer) {
	var newContainers []*typedef.ContainerInfo
	for _, container := range containers {
		newContainers = append(newContainers, container.ConvertNRIRawContainer2ContainerInfo())
	}
	manager.Pods.syncContainers2Pods(newContainers)
}

// ListOfflinePods returns offline pods
func (manager *PodManager) ListOfflinePods() ([]*typedef.PodInfo, error) {
	return nil, nil
}

// ListOnlinePods returns online pods
func (manager *PodManager) ListOnlinePods() ([]*typedef.PodInfo, error) {
	return nil, nil
}

func withOption(pi *typedef.PodInfo, opts []api.ListOption) bool {
	for _, opt := range opts {
		if !opt(pi) {
			return false
		}
	}
	return true
}

// ListContainersWithOptions filters and returns deep copy objects of all containers
func (manager *PodManager) ListContainersWithOptions(options ...api.ListOption) map[string]*typedef.ContainerInfo {
	conts := make(map[string]*typedef.ContainerInfo)
	for _, pod := range manager.ListPodsWithOptions(options...) {
		for _, ci := range pod.IDContainersMap {
			conts[ci.ID] = ci
		}
	}
	return conts
}

// ListPodsWithOptions filters and returns deep copy objects of all pods
func (manager *PodManager) ListPodsWithOptions(options ...api.ListOption) map[string]*typedef.PodInfo {
	// already deep copied
	allPods := manager.Pods.listPod()
	pods := make(map[string]*typedef.PodInfo, len(allPods))
	for _, pod := range allPods {
		if !withOption(pod, options) {
			continue
		}
		pods[pod.UID] = pod
	}
	return pods
}
