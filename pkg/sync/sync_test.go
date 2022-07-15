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
// Create: 2021-05-07
// Description: sync test

package sync

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/util"
)

// TestSync is Sync function test
func TestSync(t *testing.T) {
	os.Setenv(constant.NodeNameEnvKey, "")
	clientSet := &kubernetes.Clientset{}
	if err := Sync(true, clientSet); err != nil {
		assert.Contains(t, err.Error(), "environment variable")
	}
}

// TestGetOfflinePodStruct is getOfflinePodQosStruct function test
func TestGetOfflinePodStruct(t *testing.T) {
	pod := corev1.Pod{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			UID: "podabc",
		},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			QOSClass: corev1.PodQOSGuaranteed,
		},
	}

	cgroupPath := util.GetPodCgroupPath(&pod)
	podQosInfo, err := getOfflinePodQosStruct(string(pod.UID), cgroupPath)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/podpodabc")

	pod.Status.QOSClass = corev1.PodQOSBurstable
	cgroupPath = util.GetPodCgroupPath(&pod)
	podQosInfo, err = getOfflinePodQosStruct(string(pod.UID), cgroupPath)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/burstable/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/burstable/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/burstable/podpodabc")

	pod.Status.QOSClass = corev1.PodQOSBestEffort
	cgroupPath = util.GetPodCgroupPath(&pod)
	podQosInfo, err = getOfflinePodQosStruct(string(pod.UID), cgroupPath)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/besteffort/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/besteffort/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/besteffort/podpodabc")
}

// TestIsOffline is isOffline function test
func TestIsOffline(t *testing.T) {
	annotationMap := make(map[string]string, 1)
	annotationMap[constant.PriorityAnnotationKey] = "true"
	pod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Annotations: annotationMap,
		},
	}

	is := isOffline(pod)
	assert.Equal(t, true, is)
}

// TestVerifyPodsSetting test verifyPodsSetting
func TestVerifyPodsSetting(t *testing.T) {
	onlineAnnoMap, offlineAnnoMap, offlineAnnoMapInvalid := make(map[string]string, 0), make(map[string]string, 0),
		make(map[string]string, 0)
	onlineAnnoMap[constant.PriorityAnnotationKey], offlineAnnoMap[constant.PriorityAnnotationKey] = "false", "true"
	offlineAnnoMapInvalid[constant.PriorityAnnotationKey] = "true"
	offlineAnnoMapInvalid[cacheLimitAnnotationKey] = "invalid"
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: v1.ObjectMeta{
					UID:         "abc",
					Annotations: onlineAnnoMap,
				},
				Status: corev1.PodStatus{
					QOSClass: corev1.PodQOSGuaranteed,
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					UID:         "abc2",
					Annotations: offlineAnnoMap,
				},
				Status: corev1.PodStatus{
					QOSClass: corev1.PodQOSGuaranteed,
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					UID:         "abc3",
					Annotations: offlineAnnoMapInvalid,
				},
				Status: corev1.PodStatus{
					QOSClass: corev1.PodQOSGuaranteed,
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					UID:         "abc4",
					Annotations: offlineAnnoMap,
				},
			},
		},
	}

	verifyPodsSetting(pods, false)
	verifyPodsSetting(pods, true)
	assert.Equal(t, true, cachelimit.Init(&config.CacheConfig{}) != nil)
	verifyPodsSetting(pods, true)
}
