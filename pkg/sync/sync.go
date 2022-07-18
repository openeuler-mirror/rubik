// Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-04-20
// Description: qos setting sync

package sync

import (
	"context"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/qos"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
)

// Sync qos setting
func Sync(check bool, pods map[string]*typedef.PodInfo) error {
	for _, pod := range pods {
		if !pod.Offline {
			if pod.Namespace != "kube-system" {
				cachelimit.AddOnlinePod(pod.UID, pod.CgroupPath)
			}
			continue
		}

		if !check {
			continue
		}
		syncQos(pod.UID, pod.CgroupPath)
		if cachelimit.ClEnabled() {
			syncCache(pod)
		}
	}

	return nil
}

func syncQos(podID, cgPath string) {
	podQosInfo, err := getOfflinePodQosStruct(podID, cgPath)
	if err != nil {
		log.Errorf("get pod %v info for qos sync error: %v", podID, err)
		return
	}
	if err = podQosInfo.SetQos(); err != nil {
		log.Errorf("sync pod %v qos error: %v", podID, err)
	}
}

func syncCache(pod *typedef.PodInfo) {
	podCacheInfo, err := getCacheLimitPodStruct(pod)
	if err != nil {
		log.Errorf("get pod %v cache limit info error: %v", pod.UID, err)
		return
	}
	if err = podCacheInfo.SetCacheLimit(); err != nil {
		log.Errorf("sync pod %v cache limit error: %v", pod.UID, err)
	}
}

func getOfflinePodQosStruct(podID, cgroupPath string) (*qos.PodInfo, error) {
	podQos := api.PodQoS{
		CgroupPath: cgroupPath,
		QosLevel:   -1,
	}
	podInfo, err := qos.NewPodInfo(context.Background(), podID, config.CgroupRoot, podQos)
	if err != nil {
		return nil, err
	}

	return podInfo, nil
}

func getCacheLimitPodStruct(pod *typedef.PodInfo) (*cachelimit.PodInfo, error) {
	podQos := api.PodQoS{
		CgroupPath:      pod.CgroupPath,
		QosLevel:        -1,
		CacheLimitLevel: pod.CacheLimitLevel,
	}

	podInfo, err := cachelimit.NewCacheLimitPodInfo(context.Background(), pod.UID, podQos)
	if err != nil {
		return nil, err
	}
	return podInfo, nil
}
