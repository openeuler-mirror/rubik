// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-05-16
// Description: This file is used for expulsion trigger

package trigger

import (
	"context"
	"fmt"
	"sync"

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/informer"
)

// expulsionExec is the singleton of Expulsion executor implementation
var expulsionExec *Expulsion

// expulsionCreator creates Expulsion trigger
var expulsionCreator = func() Trigger {
	if expulsionExec == nil {
		c := newKubeClient()
		if c != nil {
			log.Infof("initialize expulsionExec")
			expulsionExec = &Expulsion{client: c}
			appendUsedExecutors(ExpulsionAnno, expulsionExec)
		}
	}
	return withTreeTirgger(ExpulsionAnno, expulsionExec)
}

// Expulsion is the trigger to evict pods
type Expulsion struct {
	sync.RWMutex
	client *kubernetes.Clientset
}

// newKubeClient returns a kubernetes.Clientset object
func newKubeClient() *kubernetes.Clientset {
	client, err := informer.InitKubeClient()
	if err != nil {
		log.Errorf("fail to connect k8s: %v", err)
		return nil
	}
	return client
}

// Execute evicts pods based on the id of the given pod
func (e *Expulsion) Execute(f Factor) (Factor, error) {
	e.RLock()
	defer e.RUnlock()
	if e.client == nil {
		return nil, fmt.Errorf("fail to use kubernetes client, please check")
	}
	var errs error
	for name, pod := range f.TargetPods() {
		log.Infof("evicting pod \"%v\"", name)
		if err := inevictable(pod); err != nil {
			errs = util.AppendErr(errs, fmt.Errorf("fail to evict pod \"%v\": %v", pod.Name, err))
			continue
		}
		eviction := &policyv1beta1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{},
		}
		if err := e.client.CoreV1().Pods(pod.Namespace).Evict(context.TODO(), eviction); err != nil {
			errs = util.AppendErr(errs, fmt.Errorf("fail to evict pod \"%v\": %v", pod.Name, err))
			continue
		}
	}
	return nil, errs
}

// Stop stops the expulsion trigger
func (e *Expulsion) Stop() error {
	e.Lock()
	defer e.Unlock()
	e.client = nil
	return nil
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
