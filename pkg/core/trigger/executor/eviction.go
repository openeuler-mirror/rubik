// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-10-31
// Description: This file is used for expulsion action

package executor

import (
	"context"
	"fmt"

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/trigger/common"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/lib/kubernetes"
)

func EvictPod(ctx context.Context) error {
	var errs error
	client, err := kubernetes.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get kubernetes client: %v", err)
	}
	pods, ok := ctx.Value(common.TARGETPODS).(map[string]*typedef.PodInfo)
	if !ok {
		return fmt.Errorf("failed to get target pods")
	}
	eviction := &policyv1beta1.Eviction{
		ObjectMeta:    metav1.ObjectMeta{},
		DeleteOptions: &metav1.DeleteOptions{},
	}
	for name, pod := range pods {
		log.Infof("evicting pod \"%v\"", name)
		if err := inevictable(pod); err != nil {
			errs = util.AppendErr(errs, fmt.Errorf("failed to evict pod \"%v\": %v", pod.Name, err))
			continue
		}
		eviction.ObjectMeta.Name = pod.Name
		eviction.ObjectMeta.Namespace = pod.Namespace
		if err := client.CoreV1().Pods(pod.Namespace).Evict(context.TODO(), eviction); err != nil {
			errs = util.AppendErr(errs, fmt.Errorf("failed to evict pod \"%v\": %v", pod.Name, err))
		}
	}
	return errs
}

func inevictable(pod *typedef.PodInfo) error {
	var forbidden = map[string]struct{}{
		"kube-system": {},
	}
	if _, existed := forbidden[pod.Namespace]; existed {
		return fmt.Errorf("it is forbidden to delete the pod whose namespace is %v", pod.Namespace)
	}
	return nil
}
