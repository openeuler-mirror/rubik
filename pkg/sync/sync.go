// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: jingxiaolu
// Create: 2021-09-30
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
	"k8s.io/client-go/rest"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/qos"
	log "isula.org/rubik/pkg/tinylog"
)

const (
	defaultNodeCgroupName = "kubepods"
	podCgroupNamePrefix   = "pod"
	nodeNameEnvKey        = "RUBIK_NODE_NAME"
	priorityAnnotationKey = "volcano.sh/preemptable"
)

// Sync qos setting
func Sync() error {
	log.Logf("Syncing qos level start")

	clientSet, err := getClient()
	if err != nil {
		return err
	}

	err = verifyOfflinePods(clientSet)

	log.Logf("Syncing qos level done")
	return err
}

func getClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func verifyOfflinePods(clientSet *kubernetes.Clientset) error {
	node := os.Getenv(nodeNameEnvKey)
	if node == "" {
		return errors.Errorf("environment variable %s must be defined", nodeNameEnvKey)
	}

	pods, err := clientSet.CoreV1().Pods("").List(context.Background(),
		metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.nodeName=%s", node)})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if !isOffline(pod) {
			continue
		}
		podQosInfo, err := getOfflinePodStruct(pod)
		if err != nil {
			log.Errorf("get pod %v info for sync error: %v", pod.UID, err)
			continue
		}
		sErr := podQosInfo.SetQos()
		log.Logf("pod %v qos level check error: %v, reset qos error: %v", pod.UID, err, sErr)
	}

	return nil
}

func isOffline(pod corev1.Pod) bool {
	return pod.Annotations[priorityAnnotationKey] == "true"
}

func getOfflinePodStruct(pod corev1.Pod) (*qos.PodInfo, error) {
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
