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
	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/qos"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
)

// Sync qos setting
func Sync(pods map[string]*typedef.PodInfo) error {
	for _, pod := range pods {
		if err := qos.SetQosLevel(pod); err != nil {
			log.Errorf("sync set pod %v qoslevel error: %v", pod.UID, err)
		}
		if cachelimit.ClEnabled() {
			syncCache(pod)
		}
	}

	return nil
}

func syncCache(pi *typedef.PodInfo) {
	err := cachelimit.SyncLevel(pi)
	if err != nil {
		log.Errorf("sync pod %v level error: %v", pi.UID, err)
		return
	}
	if err = cachelimit.SetCacheLimit(pi); err != nil {
		log.Errorf("sync pod %v cache limit error: %v", pi.UID, err)
	}
}
