// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jingxiao Lu
// Create: 2022-05-25
// Description: tests for pod.go

package util

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/constant"
)

const (
	trueStr = "true"
)

func TestIsOffline(t *testing.T) {
	var pod = &corev1.Pod{}
	pod.Annotations = make(map[string]string)
	pod.Annotations[constant.PriorityAnnotationKey] = trueStr
	if !IsOffline(pod) {
		t.Fatalf("%s failed for Annotations is %s", t.Name(), trueStr)
	}

	delete(pod.Annotations, constant.PriorityAnnotationKey)
	if IsOffline(pod) {
		t.Fatalf("%s failed for Annotations no such key", t.Name())
	}
}

// TestGetQuotaBurst is testcase for GetQuotaBurst
func TestGetQuotaBurst(t *testing.T) {
	pod := &corev1.Pod{}
	pod.Annotations = make(map[string]string)
	maxInt64PlusOne := "9223372036854775808"
	tests := []struct {
		name       string
		quotaBurst string
		want       int64
	}{
		{
			name:       "TC1-valid quota burst",
			quotaBurst: "1",
			want:       1,
		},
		{
			name:       "TC2-empty quota burst",
			quotaBurst: "",
			want:       -1,
		},
		{
			name:       "TC3-zero quota burst",
			quotaBurst: "0",
			want:       0,
		},
		{
			name:       "TC4-negative quota burst",
			quotaBurst: "-100",
			want:       -1,
		},
		{
			name:       "TC5-float quota burst",
			quotaBurst: "100.34",
			want:       -1,
		},
		{
			name:       "TC6-nonnumerical quota burst",
			quotaBurst: "nonnumerical",
			want:       -1,
		},
		{
			name:       "TC7-exceed max int64",
			quotaBurst: maxInt64PlusOne,
			want:       -1,
		},
	}
	for _, tt := range tests {
		pod.Annotations[constant.QuotaBurstAnnotationKey] = tt.quotaBurst
		assert.Equal(t, GetQuotaBurst(pod), tt.want)
	}
}

func TestGetPodCgroupPath(t *testing.T) {
	var pod = &corev1.Pod{}
	pod.UID = "AAA"
	var guaranteedPath = filepath.Join(constant.KubepodsCgroup, constant.PodCgroupNamePrefix+string(pod.UID))
	var burstablePath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)), constant.PodCgroupNamePrefix+string(pod.UID))
	var besteffortPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)), constant.PodCgroupNamePrefix+string(pod.UID))
	pod.Annotations = make(map[string]string)

	// no pod.Annotations[configHashAnnotationKey]
	pod.Status.QOSClass = corev1.PodQOSGuaranteed
	if !assert.Equal(t, GetPodCgroupPath(pod), guaranteedPath) {
		t.Fatalf("%s failed for PodQOSGuaranteed without configHash", t.Name())
	}
	pod.Status.QOSClass = corev1.PodQOSBurstable
	if !assert.Equal(t, GetPodCgroupPath(pod), burstablePath) {
		t.Fatalf("%s failed for PodQOSBurstable without configHash", t.Name())
	}
	pod.Status.QOSClass = corev1.PodQOSBestEffort
	if !assert.Equal(t, GetPodCgroupPath(pod), besteffortPath) {
		t.Fatalf("%s failed for PodQOSBestEffort without configHash", t.Name())
	}
	pod.Status.QOSClass = ""
	if !assert.Equal(t, GetPodCgroupPath(pod), "") {
		t.Fatalf("%s failed for not setting QOSClass without configHash", t.Name())
	}

	// has pod.Annotations[configHashAnnotationKey]
	pod.Annotations[configHashAnnotationKey] = "BBB"
	var id = pod.Annotations[configHashAnnotationKey]
	guaranteedPath = filepath.Join(constant.KubepodsCgroup, constant.PodCgroupNamePrefix+id)
	burstablePath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)), constant.PodCgroupNamePrefix+id)
	besteffortPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)), constant.PodCgroupNamePrefix+id)
	pod.Status.QOSClass = corev1.PodQOSGuaranteed
	if !assert.Equal(t, GetPodCgroupPath(pod), guaranteedPath) {
		t.Fatalf("%s failed for PodQOSGuaranteed with configHash", t.Name())
	}
	pod.Status.QOSClass = corev1.PodQOSBurstable
	if !assert.Equal(t, GetPodCgroupPath(pod), burstablePath) {
		t.Fatalf("%s failed for PodQOSBurstable with configHash", t.Name())
	}
	pod.Status.QOSClass = corev1.PodQOSBestEffort
	if !assert.Equal(t, GetPodCgroupPath(pod), besteffortPath) {
		t.Fatalf("%s failed for PodQOSBestEffort with configHash", t.Name())
	}
	pod.Status.QOSClass = ""
	if !assert.Equal(t, GetPodCgroupPath(pod), "") {
		t.Fatalf("%s failed for not setting QOSClass with configHash", t.Name())
	}
}
