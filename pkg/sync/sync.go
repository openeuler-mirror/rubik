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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"isula.org/rubik/pkg/common"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
)

// Sync qos setting
func Sync(kubeClient *kubernetes.Clientset) error {
	log.Logf("Syncing qos level start")

	node := os.Getenv(constant.NodeNameEnvKey)
	if node == "" {
		return errors.Errorf("environment variable %s must be defined", constant.NodeNameEnvKey)
	}

	pods, err := kubeClient.CoreV1().Pods("").List(context.Background(),
		metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.nodeName=%s", node)})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if !common.IsOffline(pod) {
			continue
		}
		podQosInfo, err := common.BuildOfflinePodInfo(pod)
		if err != nil {
			log.Errorf("get pod %v info for sync error: %v", pod.UID, err)
			continue
		}
		if err = podQosInfo.ValidateQos(); err != nil {
			sErr := podQosInfo.SetQos()
			log.Logf("pod %v qos level check error: %v, reset qos error: %v", pod.UID, err, sErr)
		}
	}

	log.Logf("Syncing qos level done")

	return nil
}
