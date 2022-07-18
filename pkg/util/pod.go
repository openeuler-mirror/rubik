// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2022-05-25
// Description: Pod related common functions

package util

import (
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/constant"
)

const (
	configHashAnnotationKey = "kubernetes.io/config.hash"
)

// IsOffline judges whether pod is offline pod
func IsOffline(pod *corev1.Pod) bool {
	return pod.Annotations[constant.PriorityAnnotationKey] == "true"
}

func GetPodCacheLimit(pod *corev1.Pod) string {
	return pod.Annotations[constant.CacheLimitAnnotationKey]
}

// GetPodCgroupPath returns cgroup path of pod
func GetPodCgroupPath(pod *corev1.Pod) string {
	var cgroupPath string
	id := string(pod.UID)
	if configHash := pod.Annotations[configHashAnnotationKey]; configHash != "" {
		id = configHash
	}

	switch pod.Status.QOSClass {
	case corev1.PodQOSGuaranteed:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, constant.PodCgroupNamePrefix+id)
	case corev1.PodQOSBurstable:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)),
			constant.PodCgroupNamePrefix+id)
	case corev1.PodQOSBestEffort:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)),
			constant.PodCgroupNamePrefix+id)
	}

	return cgroupPath
}
