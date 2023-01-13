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

// podCache is used to store PodInfo
type podCache struct {
	sync.RWMutex
	Pods map[string]*typedef.PodInfo
}

// NewPodCache returns a PodCache object (pointer)
func NewPodCache() *podCache {
	return &podCache{
		Pods: make(map[string]*typedef.PodInfo, 0),
	}
}

// getPod returns the deepcopy object of pod
func (cache *podCache) getPod(podID string) *typedef.PodInfo {
	cache.RLock()
	defer cache.RUnlock()
	return cache.Pods[podID].Clone()
}

// podExist returns true if there is a pod whose key is podID in the pods
func (cache *podCache) podExist(podID string) bool {
	cache.RLock()
	_, ok := cache.Pods[podID]
	cache.RUnlock()
	return ok
}

// addPod adds pod information
func (cache *podCache) addPod(pod *typedef.PodInfo) {
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
func (cache *podCache) delPod(podID string) {
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
func (cache *podCache) updatePod(pod *typedef.PodInfo) {
	if pod == nil || pod.UID == "" {
		return
	}
	cache.Lock()
	cache.Pods[pod.UID] = pod
	cache.Unlock()
	log.Debugf("update pod %v", pod.UID)
}
