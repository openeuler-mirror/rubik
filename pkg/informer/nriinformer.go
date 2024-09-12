// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: weiyuan
// Create: 2024-05-28
// Description: This file defines nriinformer which interact with nri

// Package informer implements informer interface
package informer

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"

	rubikapi "isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
)

const (
	nriPluginName  = "rubik"
	nriPluginIndex = "10"
)

// NRIInformer interacts with nri server and forward data to the internal
type NRIInformer struct {
	rubikapi.Publisher
	nodeName     string
	stub         stub.Stub
	finishedSync chan struct{}
}

// NewNRIInformer create an rubik nri plugin
func NewNRIInformer(publisher rubikapi.Publisher) (rubikapi.Informer, error) {
	p := &NRIInformer{
		Publisher:    publisher,
		nodeName:     os.Getenv(constant.NodeNameEnvKey),
		finishedSync: make(chan struct{}),
	}

	options := []stub.Option{
		stub.WithPluginName(nriPluginName),
		stub.WithPluginIdx(nriPluginIndex),
	}
	s, err := stub.New(p, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create stub: %v", err)
	}
	p.stub = s
	return p, err
}

// Start starts nri informer
func (plugin NRIInformer) Start(ctx context.Context) error {
	if err := plugin.stub.Start(ctx); err != nil {
		return fmt.Errorf("failed to start nri informer: %v", err)
	}
	<-plugin.finishedSync

	go func() {
		plugin.stub.Wait()
		plugin.stub.Stop()
	}()
	return nil
}

// Synchronize syncs the nri containers & sandboxes
func (plugin NRIInformer) Synchronize(ctx context.Context, pods []*api.PodSandbox, containers []*api.Container) (
	[]*api.ContainerUpdate, error) {
	plugin.Publish(typedef.NRIPODSYNCALL, pods)
	plugin.Publish(typedef.NRICONTAINERSYNCALL, containers)
	// notify service handler to start
	close(plugin.finishedSync)
	return nil, nil
}

// RunPodSandbox will be called when sandbox starts.
func (plugin NRIInformer) RunPodSandbox(ctx context.Context, pod *api.PodSandbox) error {
	plugin.Publish(typedef.NRIPODADD, pod)
	return nil
}

// StopPodSandbox will be called when sandbox stops.
func (plugin NRIInformer) StopPodSandbox(ctx context.Context, pod *api.PodSandbox) error {
	return nil
}

// RemovePodSandbox will be called when sandbox is removed.
func (plugin NRIInformer) RemovePodSandbox(ctx context.Context, pod *api.PodSandbox) error {
	plugin.Publish(typedef.NRIPODDELETE, pod)
	return nil
}

// CreateContainer will be called when it creates container
func (plugin NRIInformer) CreateContainer(context.Context, *api.PodSandbox, *api.Container) (
	*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	return nil, nil, nil
}

// StartContainer will be called when container starts
func (plugin NRIInformer) StartContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	plugin.Publish(typedef.NRICONTAINERSTART, container)
	return nil
}

// UpdateContainer will be called when container updates
func (plugin NRIInformer) UpdateContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container, lr *api.LinuxResources) ([]*api.ContainerUpdate, error) {
	return nil, nil
}

// StopContainer will be called when container stops
func (plugin NRIInformer) StopContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container) ([]*api.ContainerUpdate, error) {
	plugin.Publish(typedef.NRICONTAINERREMOVE, container)
	return nil, nil
}

// RemoveContainer will be called when it removes container
func (plugin NRIInformer) RemoveContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	plugin.Publish(typedef.NRICONTAINERREMOVE, container)
	return nil
}

// PostCreateContainer will be called after container was created
func (plugin NRIInformer) PostCreateContainer(context.Context, *api.PodSandbox, *api.Container) error {
	return nil
}

// ostStartContainer will be called after container was started
func (plugin NRIInformer) PostStartContainer(context.Context, *api.PodSandbox, *api.Container) error {
	return nil
}

// PostUpdateContainer will be called after container was updated
func (plugin NRIInformer) PostUpdateContainer(context.Context, *api.PodSandbox, *api.Container) error {
	return nil
}

// Shutdown will be called when nri plugin shutdowns
func (plugin NRIInformer) Shutdown(context.Context) {
}
