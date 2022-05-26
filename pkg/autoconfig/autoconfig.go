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
// Description: qos auto config

// Package autoconfig is for qos auto config
package autoconfig

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
)

// EventHandler is used to process pod events pushed by Kubernetes APIServer.
type EventHandler interface {
	AddEvent(pod *corev1.Pod)
	UpdateEvent(oldPod *corev1.Pod, newPod *corev1.Pod)
	DeleteEvent(pod *corev1.Pod)
}

// Backend is Rubik struct.
var Backend EventHandler

// Init initializes the callback function for the pod event.
func Init(kubeClient *kubernetes.Clientset) error {
	const (
		reSyncTime        = 30
		specNodeNameField = "spec.nodeName"
	)
	nodeName := os.Getenv(constant.NodeNameEnvKey)
	if nodeName == "" {
		return fmt.Errorf("environment variable %s must be defined", constant.NodeNameEnvKey)
	}
	kubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient,
		time.Duration(reSyncTime)*time.Second,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			// set Options to return only pods on the current node.
			options.FieldSelector = fields.OneTermEqualSelector(specNodeNameField, nodeName).String()
		}))
	kubeInformerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addHandler,
		UpdateFunc: updateHandler,
		DeleteFunc: deleteHandler,
	})
	kubeInformerFactory.Start(config.ShutdownChan)
	return nil
}

func updateHandler(old, new interface{}) {
	oldPod, ok1 := old.(*corev1.Pod)
	newPod, ok2 := new.(*corev1.Pod)
	if !ok1 || !ok2 {
		log.Errorf("auto config error: invalid pod type")
		return
	}

	if !util.IsOffline(newPod) || newPod.Status.Phase != "Running" {
		return
	}

	Backend.UpdateEvent(oldPod, newPod)
}

func addHandler(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.Errorf("auto config error: invalid pod type")
		return
	}

	if !util.IsOffline(pod) || !isPodOnCurrentNode(pod) {
		return
	}

	Backend.AddEvent(pod)
}

func deleteHandler(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.Errorf("invalid pod type")
		return
	}

	Backend.DeleteEvent(pod)
}

func isPodOnCurrentNode(pod *corev1.Pod) bool {
	currentNode := os.Getenv(constant.NodeNameEnvKey)
	if currentNode == "" {
		log.Errorf("auto config error: environment variable %s must be defined", constant.NodeNameEnvKey)
		return false
	}

	return pod.Spec.NodeName == currentNode
}
