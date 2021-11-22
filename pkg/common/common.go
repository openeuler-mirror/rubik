// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-07-22
// Description: common functions

package common

import (
	"context"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/qos"
)

const (
	defaultNodeCgroupName = "kubepods"
	podCgroupNamePrefix   = "pod"
	priorityAnnotationKey = "volcano.sh/preemptable"
)

// IsOffline judges whether pod is offline pod
func IsOffline(pod corev1.Pod) bool {
	return pod.Annotations[priorityAnnotationKey] == "true"
}

// BuildOfflinePodInfo build offline pod information
func BuildOfflinePodInfo(pod corev1.Pod) (*qos.PodInfo, error) {
	var cgroupPath string
	switch pod.Status.QOSClass {
	case corev1.PodQOSGuaranteed:
		cgroupPath = filepath.Join(defaultNodeCgroupName, podCgroupNamePrefix+string(pod.UID))
	case corev1.PodQOSBurstable:
		cgroupPath = filepath.Join(defaultNodeCgroupName, strings.ToLower(string(corev1.PodQOSBurstable)),
			podCgroupNamePrefix+string(pod.UID))
	case corev1.PodQOSBestEffort:
		cgroupPath = filepath.Join(defaultNodeCgroupName, strings.ToLower(string(corev1.PodQOSBestEffort)),
			podCgroupNamePrefix+string(pod.UID))
	}

	podQos := api.PodQoS{
		CgroupPath: cgroupPath,
		QosLevel:   -1,
	}
	podInfo, err := qos.NewPodInfo(context.Background(), string(pod.UID), config.CgroupRoot, podQos)
	if err != nil {
		return nil, err
	}

	return podInfo, nil
}
