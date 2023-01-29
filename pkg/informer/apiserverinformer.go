// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2023-01-05
// Description: This file defines apiinformer which interact with kubernetes apiserver

// Package informer implements informer interface
package informer

import (
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
)

// APIServerInformer interacts with k8s api server and forward data to the internal
type APIServerInformer struct {
	api.Publisher
	client   *kubernetes.Clientset
	nodeName string
}

// NewAPIServerInformer creates an PIServerInformer instance
func NewAPIServerInformer(publisher api.Publisher) (api.Informer, error) {
	informer := &APIServerInformer{
		Publisher: publisher,
	}

	// create apiserver client
	client, err := initKubeClient()
	if err != nil {
		return nil, fmt.Errorf("fail to init kubenetes client: %v", err)
	}
	informer.client = client

	// filter pods on current nodes
	nodeName := os.Getenv(constant.NodeNameEnvKey)
	if nodeName == "" {
		return nil, fmt.Errorf("missing %s", constant.NodeNameEnvKey)
	}
	informer.nodeName = nodeName

	return informer, nil
}

// initKubeClient initializes kubeClient
func initKubeClient() (*kubernetes.Clientset, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return kubeClient, nil
}

// Start starts and enables PIServerInformer
func (ai *APIServerInformer) Start(stopCh <-chan struct{}) {
	const (
		reSyncTime        = 30
		specNodeNameField = "spec.nodeName"
	)
	kubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(ai.client,
		time.Duration(reSyncTime)*time.Second,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			// set Options to return only pods on the current node.
			options.FieldSelector = fields.OneTermEqualSelector(specNodeNameField, ai.nodeName).String()
		}))
	kubeInformerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ai.addFunc,
		UpdateFunc: ai.updateFunc,
		DeleteFunc: ai.deleteFunc,
	})
	kubeInformerFactory.Start(stopCh)
}

func (ai *APIServerInformer) addFunc(obj interface{}) {
	ai.Publish(typedef.RAW_POD_ADD, obj)
}

func (ai *APIServerInformer) updateFunc(oldObj, newObj interface{}) {
	ai.Publish(typedef.RAW_POD_UPDATE, newObj)
}

func (ai *APIServerInformer) deleteFunc(obj interface{}) {
	ai.Publish(typedef.RAW_POD_DELETE, obj)
}