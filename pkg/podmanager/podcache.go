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
// Description: This file defines pod cache storing pod information

// Package podmanager implements cache connecting informer and module manager
package podmanager

import (
	"sync"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
)

// PodCache is used to store PodInfo
type PodCache struct {
	sync.RWMutex
	Pods map[string]*typedef.PodInfo
}

// NewPodCache returns a PodCache object (pointer)
func NewPodCache() *PodCache {
	return &PodCache{
		Pods: make(map[string]*typedef.PodInfo, 0),
	}
}

// getPod returns the deepcopy object of pod
func (cache *PodCache) getPod(podID string) *typedef.PodInfo {
	cache.RLock()
	defer cache.RUnlock()
	return cache.Pods[podID].DeepCopy()
}

// podExist returns true if there is a pod whose key is podID in the pods
func (cache *PodCache) podExist(podID string) bool {
	cache.RLock()
	_, ok := cache.Pods[podID]
	cache.RUnlock()
	return ok
}

// addPod adds pod information
func (cache *PodCache) addPod(pod *typedef.PodInfo) {
	if pod == nil || pod.UID == "" {
		return
	}
	if ok := cache.podExist(pod.UID); ok {
		log.Debugf("pod %v is existed", string(pod.UID))
		return
	}
	cache.Lock()
	cache.Pods[pod.UID] = pod
	cache.Unlock()
	log.Debugf("add pod %v", string(pod.UID))
}

// delPod deletes pod information
func (cache *PodCache) delPod(podID string) {
	if ok := cache.podExist(podID); !ok {
		log.Debugf("pod %v is not existed", string(podID))
		return
	}
	cache.Lock()
	delete(cache.Pods, podID)
	cache.Unlock()
	log.Debugf("delete pod %v", podID)
}

// updatePod updates pod information
func (cache *PodCache) updatePod(pod *typedef.PodInfo) {
	if pod == nil || pod.UID == "" {
		return
	}
	cache.Lock()
	cache.Pods[pod.UID] = pod
	cache.Unlock()
	log.Debugf("update pod %v", pod.UID)
}

// substitute replaces all the data in the cache
func (cache *PodCache) substitute(pods []*typedef.PodInfo) {
	cache.Lock()
	defer cache.Unlock()
	cache.Pods = make(map[string]*typedef.PodInfo, 0)
	if len(pods) == 0 {
		return
	}
	for _, pod := range pods {
		if pod == nil || pod.UID == "" {
			continue
		}
		cache.Pods[pod.UID] = pod
		log.Debugf("substituting pod %v", pod.UID)
	}
}

// listPod returns the deepcopy object of all pod
func (cache *PodCache) listPod() map[string]*typedef.PodInfo {
	res := make(map[string]*typedef.PodInfo, len(cache.Pods))
	cache.RLock()
	for id, pi := range cache.Pods {
		res[id] = pi.DeepCopy()
	}
	cache.RUnlock()
	return res
}
