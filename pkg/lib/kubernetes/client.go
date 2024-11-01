// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
//	http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-10-31
// Description: This file is used for kubernetes client

package kubernetes

import (
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	kubernetes.Clientset
}

var (
	defaultClient *Client
	clientSync    sync.RWMutex
)

// initKubeClient initializes kubeClient
func initClient() (*Client, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return &Client{Clientset: *kubeClient}, nil
}

// GetClient gets the globally unique default kubernetes client
func GetClient() (*Client, error) {
	// prevent multiple initializations
	clientSync.Lock()
	defer clientSync.Unlock()

	if defaultClient != nil {
		return defaultClient, nil
	}
	c, err := initClient()
	if err != nil {
		return nil, fmt.Errorf("failed to init client: %v", err)
	}
	defaultClient = c
	return defaultClient, nil
}
