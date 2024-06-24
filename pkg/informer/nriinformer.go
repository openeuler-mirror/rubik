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
	"os"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	rubikapi "isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
)

// NRIInformer interacts with crio server and forward data to the internal
type NRIInformer struct {
	rubikapi.Publisher
	nodeName     string
	stub         stub.Stub
	finishedSync chan struct{}
}

// register nriplugin
func NewNRIInformer(publisher rubikapi.Publisher) (rubikapi.Informer, error) {
	p := &NRIInformer{}
	p.Publisher = publisher
	pluginName := "rubik"
	pluginIndex := "10"
	nodeName := os.Getenv(constant.NodeNameEnvKey)
	p.finishedSync = make(chan struct{})
	p.nodeName = nodeName
	var err error
	options := []stub.Option{
		stub.WithPluginName(pluginName),
		stub.WithPluginIdx(pluginIndex),
	}
	p.stub, err = stub.New(p, options...)
	return p, err
}

// start nriplugin
func (plugin NRIInformer) Start(ctx context.Context) {
	go plugin.stub.Run(ctx)
}

// wait sync event finish
func (plugin NRIInformer) WaitReady() {
	<-plugin.finishedSync
}

// nri sync event
func (plugin NRIInformer) Synchronize(ctx context.Context, pods []*api.PodSandbox, containers []*api.Container) ([]*api.ContainerUpdate, error) {
	plugin.Publish(typedef.NRIPODSYNCALL, pods)
	plugin.Publish(typedef.NRICONTAINERSYNCALL, containers)
	// notify service handler to start
	close(plugin.finishedSync)

	return nil, nil
}

// nri pod run event
func (plugin NRIInformer) RunPodSandbox(ctx context.Context, pod *api.PodSandbox) error {
	plugin.Publish(typedef.NRIPODADD, pod)
	return nil
}

// nri pod stop event
func (plugin NRIInformer) StopPodSandbox(ctx context.Context, pod *api.PodSandbox) error {
	return nil
}

// nri pod remove event
func (plugin NRIInformer) RemovePodSandbox(ctx context.Context, pod *api.PodSandbox) error {
	plugin.Publish(typedef.NRIPODDELETE, pod)

	return nil
}

// nri container create event
func (plugin NRIInformer) CreateContainer(context.Context, *api.PodSandbox, *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	var containerAdjustment = &api.ContainerAdjustment{}
	var containerUpdate = []*api.ContainerUpdate{}
	return containerAdjustment, containerUpdate, nil
}

// nri container start event
func (plugin NRIInformer) StartContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	plugin.Publish(typedef.NRICONTAINERSTART, container)
	return nil
}

// nri container update event
func (plugin NRIInformer) UpdateContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container, lr *api.LinuxResources) ([]*api.ContainerUpdate, error) {
	var containerUpdate = []*api.ContainerUpdate{}
	return containerUpdate, nil
}

// nri container stop cont
func (plugin NRIInformer) StopContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container) ([]*api.ContainerUpdate, error) {
	plugin.Publish(typedef.NRICONTAINERREMOVE, container)
	var containerUpdate = []*api.ContainerUpdate{}
	return containerUpdate, nil
}

// nri container remove event
func (plugin NRIInformer) RemoveContainer(ctx context.Context, pod *api.PodSandbox, container *api.Container) error {
	plugin.Publish(typedef.NRICONTAINERREMOVE, container)
	return nil
}

// nri configure event
func (plugin NRIInformer) Configure(context.Context, string, string, string) (api.EventMask, error) {
	var eventMask api.EventMask
	return eventMask, nil
}

// nri post container create event
func (plugin NRIInformer) PostCreateContainer(context.Context, *api.PodSandbox, *api.Container) error {
	return nil
}

// nri post container start event
func (plugin NRIInformer) PostStartContainer(context.Context, *api.PodSandbox, *api.Container) error {
	return nil
}

// nri post container update event
func (plugin NRIInformer) PostUpdateContainer(context.Context, *api.PodSandbox, *api.Container) error {
	return nil
}

// nri shutdown event
func (plugin NRIInformer) Shutdown(context.Context) {
}
