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
// Create: 2023-01-28
// Description: This file defines rubik agent to control the life cycle of each component

// // Package rubik defines the overall logic
package rubik

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/core/publisher"
	"isula.org/rubik/pkg/informer"
	"isula.org/rubik/pkg/podmanager"
	"isula.org/rubik/pkg/services"
)

// Agent runs a series of rubik services and manages data
type Agent struct {
	Config     *config.Config
	PodManager *podmanager.PodManager
	stopCh     *util.DisposableChannel
}

// NewAgent returns an agent for given configuration
func NewAgent(cfg *config.Config) *Agent {
	publisher := publisher.GetPublisherFactory().GetPublisher(publisher.TYPE_GENERIC)
	a := &Agent{
		Config:     cfg,
		stopCh:     util.NewDisposableChannel(),
		PodManager: podmanager.NewPodManager(publisher),
	}
	return a
}

// Run starts and runs the agent until receiving stop signal
func (a *Agent) Run() error {
	go a.handleSignals()
	if err := a.startServiceHandler(); err != nil {
		return err
	}
	if err := a.startInformer(); err != nil {
		return err
	}
	a.stopCh.Wait()
	if err := a.stopServiceHandler(); err != nil {
		return err
	}
	return nil
}

// handleSignals handles external signal input
func (a *Agent) handleSignals() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	for sig := range signalChan {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			log.Infof("signal %v received and starting exit...", sig)
			a.stopCh.Close()
		}
	}
}

// startInformer starts informer to obtain external data
func (a *Agent) startInformer() error {
	publisher := publisher.GetPublisherFactory().GetPublisher(publisher.TYPE_GENERIC)
	informer, err := informer.GetInfomerFactory().GetInformerCreator(informer.APISERVER)(publisher)
	if err != nil {
		return fmt.Errorf("fail to set informer: %v", err)
	}
	informer.Start(a.stopCh.Channel())
	return nil
}

// startServiceHandler starts and runs the service
func (a *Agent) startServiceHandler() error {
	serviceManager := services.GetServiceManager()
	if err := serviceManager.Setup(a.PodManager); err != nil {
		return fmt.Errorf("error setting service handler: %v", err)
	}
	if err := serviceManager.Start(a.stopCh.Channel()); err != nil {
		return fmt.Errorf("error setting service handler: %v", err)
	}
	a.PodManager.Subscribe(serviceManager)
	return nil
}

// stopServiceHandler stops the service
func (a *Agent) stopServiceHandler() error {
	serviceManager := services.GetServiceManager()
	if err := serviceManager.Stop(); err != nil {
		return fmt.Errorf("error stop service handler: %v", err)
	}
	return nil
}

// RunAgent creates and runs rubik's agent
func RunAgent() int {
	c := config.NewConfig(config.JSON)
	if err := c.LoadConfig(constant.ConfigFile); err != nil {
		log.Errorf("load config failed: %v\n", err)
		return -1
	}
	agent := NewAgent(c)
	if err := agent.Run(); err != nil {
		log.Errorf("error running agent: %v", err)
		return -1
	}
	return 0
}
