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
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/qos"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
)

// InitAutoConfig init qos auto config handler
func InitAutoConfig(kubeClient *kubernetes.Clientset) {
	log.Logf("qos auto config init start")

	reSyncTime := 30
	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, time.Duration(reSyncTime)*time.Second)
	kubeInformerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addHandler,
		UpdateFunc: updateHandler,
	})
	stopCh := make(chan struct{})
	kubeInformerFactory.Start(stopCh)

	log.Logf("qos auto config init success")
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

	var (
		judge1 = oldPod.Status.Phase == newPod.Status.Phase
		judge2 = oldPod.Spec.NodeName == newPod.Spec.NodeName
		judge3 = util.IsOffline(oldPod) == util.IsOffline(newPod)
	)
	// qos related status no difference, just return
	if judge1 && judge2 && judge3 {
		return
	}

	addHandler(new)
}

func addHandler(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.Errorf("auto config error: invalid pod type")
		return
	}

	node := os.Getenv(constant.NodeNameEnvKey)
	if node == "" {
		log.Errorf("auto config error: environment variable %s must be defined", constant.NodeNameEnvKey)
		return
	}
	if (pod.Spec.NodeName != node) || !util.IsOffline(pod) {
		return
	}

	if pod.Status.Phase == "Running" {
		podQosInfo, err := qos.BuildOfflinePodInfo(pod)
		if err != nil {
			log.Errorf("get pod %v info for auto config error: %v", pod.UID, err)
			return
		}
		if err := podQosInfo.SetQos(); err != nil {
			log.Errorf("auto config qos error: %v", err)
			return
		}
	}
}
