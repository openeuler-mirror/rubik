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
	"context"
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
	"isula.org/rubik/pkg/common/log"
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
	client, err := InitKubeClient()
	if err != nil {
		return nil, fmt.Errorf("failed to init kubenetes client: %v", err)
	}
	informer.client = client

	// filter pods on current nodes
	nodeName := os.Getenv(constant.NodeNameEnvKey)
	if nodeName == "" {
		return nil, fmt.Errorf("unable to get node name from environment variable %s", constant.NodeNameEnvKey)
	}
	informer.nodeName = nodeName

	return informer, nil
}

// InitKubeClient initializes kubeClient
func InitKubeClient() (*kubernetes.Clientset, error) {
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
func (informer *APIServerInformer) Start(ctx context.Context) {
	const specNodeNameField = "spec.nodeName"
	// set options to return only pods on the current node.
	var fieldSelector = fields.OneTermEqualSelector(specNodeNameField, informer.nodeName).String()
	informer.listFunc(fieldSelector)
	informer.watchFunc(ctx, fieldSelector)
}

func (informer *APIServerInformer) listFunc(fieldSelector string) {
	pods, err := informer.client.CoreV1().Pods("").List(context.Background(),
		metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		log.Errorf("failed to get pod list from APIServer informer: %v", err)
		return
	}
	informer.Publish(typedef.RAWPODSYNCALL, pods.Items)
}

func (informer *APIServerInformer) watchFunc(ctx context.Context, fieldSelector string) {
	const reSyncTime = 30
	kubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(informer.client,
		time.Duration(reSyncTime)*time.Second,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fieldSelector
		}))
	kubeInformerFactory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    informer.AddFunc,
		UpdateFunc: informer.UpdateFunc,
		DeleteFunc: informer.DeleteFunc,
	})
	kubeInformerFactory.Start(ctx.Done())
}

// AddFunc handles the raw pod increase event
func (informer *APIServerInformer) AddFunc(obj interface{}) {
	informer.Publish(typedef.RAWPODADD, obj)
}

// UpdateFunc handles the raw pod update event
func (informer *APIServerInformer) UpdateFunc(oldObj, newObj interface{}) {
	informer.Publish(typedef.RAWPODUPDATE, newObj)
}

// DeleteFunc handles the raw pod deletion event
func (informer *APIServerInformer) DeleteFunc(obj interface{}) {
	informer.Publish(typedef.RAWPODDELETE, obj)
}
