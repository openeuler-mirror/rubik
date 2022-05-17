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
// Create: 2021-04-20
// Description: qos setting sync

package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/qos"
	log "isula.org/rubik/pkg/tinylog"
)

const (
	podCgroupNamePrefix     = "pod"
	nodeNameEnvKey          = "RUBIK_NODE_NAME"
	priorityAnnotationKey   = "volcano.sh/preemptable"
	cacheLimitAnnotationKey = "volcano.sh/cache-limit"
	configHashAnnotationKey = "kubernetes.io/config.hash"
)

// Sync qos setting
func Sync(check bool, kubeClient *kubernetes.Clientset) error {
	node := os.Getenv(nodeNameEnvKey)
	if node == "" {
		return errors.Errorf("environment variable %s must be defined", nodeNameEnvKey)
	}
	pods, err := kubeClient.CoreV1().Pods("").List(context.Background(),
		metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.nodeName=%s", node)})
	if err != nil {
		return err
	}

	verifyPodsSetting(pods, check)
	return nil
}

func verifyPodsSetting(pods *corev1.PodList, check bool) {
	for _, pod := range pods.Items {
		podCgroupPath := getPodCgroupPath(pod)
		if !isOffline(pod) {
			if pod.Namespace != "kube-system" {
				cachelimit.AddOnlinePod(string(pod.UID), podCgroupPath)
			}
			continue
		}

		if !check {
			continue
		}
		syncQos(string(pod.UID), podCgroupPath)
		if cachelimit.ClEnabled() {
			syncCache(pod, podCgroupPath)
		}
	}
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

func syncCache(pod corev1.Pod, cgPath string) {
	podCacheInfo, err := getCacheLimitPodStruct(pod, cgPath)
	if err != nil {
		log.Errorf("get pod %v cache limit info error: %v", pod.UID, err)
		return
	}
	if err = podCacheInfo.SetCacheLimit(); err != nil {
		log.Errorf("sync pod %v cache limit error: %v", pod.UID, err)
	}
}

func getPodCacheLimitLevel(pod corev1.Pod) string {
	return pod.Annotations[cacheLimitAnnotationKey]
}

func isOffline(pod corev1.Pod) bool {
	return pod.Annotations[priorityAnnotationKey] == "true"
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

func getCacheLimitPodStruct(pod corev1.Pod, cgroupPath string) (*cachelimit.PodInfo, error) {
	podQos := api.PodQoS{
		CgroupPath:      cgroupPath,
		QosLevel:        -1,
		CacheLimitLevel: getPodCacheLimitLevel(pod),
	}

	podInfo, err := cachelimit.NewCacheLimitPodInfo(context.Background(), string(pod.UID), podQos)
	if err != nil {
		return nil, err
	}
	return podInfo, nil
}

func getPodCgroupPath(pod corev1.Pod) string {
	var cgroupPath string
	id := string(pod.UID)
	if configHash := pod.Annotations[configHashAnnotationKey]; configHash != "" {
		id = configHash
	}

	switch pod.Status.QOSClass {
	case corev1.PodQOSGuaranteed:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, podCgroupNamePrefix+id)
	case corev1.PodQOSBurstable:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBurstable)),
			podCgroupNamePrefix+id)
	case corev1.PodQOSBestEffort:
		cgroupPath = filepath.Join(constant.KubepodsCgroup, strings.ToLower(string(corev1.PodQOSBestEffort)),
			podCgroupNamePrefix+id)
	}

	return cgroupPath
}
