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

// Package rubik defines the overall logic
package rubik

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"

	"isula.org/rubik/pkg/api"
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
	informer   api.Informer
}

// NewAgent returns an agent for given configuration
func NewAgent(cfg *config.Config) *Agent {
	publisher := publisher.GetPublisherFactory().GetPublisher(publisher.TYPE_GENERIC)
	a := &Agent{
		Config:     cfg,
		PodManager: podmanager.NewPodManager(publisher),
	}
	return a
}

// Run starts and runs the agent until receiving stop signal
func (a *Agent) Run(ctx context.Context) error {
	if err := a.startServiceHandler(ctx); err != nil {
		return err
	}
	if err := a.startInformer(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	a.stopInformer()
	a.stopServiceHandler()
	return nil
}

// startInformer starts informer to obtain external data
func (a *Agent) startInformer(ctx context.Context) error {
	publisher := publisher.GetPublisherFactory().GetPublisher(publisher.TYPE_GENERIC)
	informer, err := informer.GetInfomerFactory().GetInformerCreator(informer.APISERVER)(publisher)
	if err != nil {
		return fmt.Errorf("fail to set informer: %v", err)
	}
	if err := informer.Subscribe(a.PodManager); err != nil {
		return fmt.Errorf("fail to subscribe informer: %v", err)
	}
	a.informer = informer
	informer.Start(ctx)
	return nil
}

// stopInformer stops the infomer
func (a *Agent) stopInformer() {
	a.informer.Unsubscribe(a.PodManager)
}

// startServiceHandler starts and runs the service
func (a *Agent) startServiceHandler(ctx context.Context) error {
	serviceManager := services.GetServiceManager()
	if err := serviceManager.Setup(a.PodManager); err != nil {
		return fmt.Errorf("error setting service handler: %v", err)
	}
	serviceManager.Start(ctx)
	a.PodManager.Subscribe(serviceManager)
	return nil
}

// stopServiceHandler stops sending data to the ServiceManager
func (a *Agent) stopServiceHandler() {
	serviceManager := services.GetServiceManager()
	a.PodManager.Unsubscribe(serviceManager)
	serviceManager.Stop()
}

// runAgent creates and runs rubik's agent
func runAgent(ctx context.Context) error {
	c := config.NewConfig(config.JSON)
	if err := c.LoadConfig(constant.ConfigFile); err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	// Agent parameter enable
	if err := log.InitConfig(c.Agent.LogDriver, c.Agent.LogDir, c.Agent.LogLevel, c.Agent.LogSize); err != nil {
		return fmt.Errorf("error initializing log: %v", err)
	}

	util.CgroupRoot = c.Agent.CgroupRoot
	agent := NewAgent(c)
	if err := agent.Run(ctx); err != nil {
		return fmt.Errorf("error running agent: %v", err)
	}
	return nil
}

// Run runs agent and process signal
func Run() int {
	// 0. file mask permission setting and parameter checking
	unix.Umask(constant.DefaultUmask)
	if len(os.Args) > 1 {
		fmt.Println("args not allowed")
		return constant.ArgumentErrorExitCode
	}
	// 1. apply file locks, only one rubik process can run at the same time
	lock, err := util.LockFile(constant.LockFile)
	defer func() {
		lock.Close()
		os.Remove(constant.LockFile)
	}()
	if err != nil {
		fmt.Printf("set rubik lock failed: %v, check if there is another rubik running\n", err)
		return constant.RepeatRunExitCode
	}
	defer util.UnlockFile(lock)

	ctx, cancel := context.WithCancel(context.Background())
	// 2. handle external signals
	go handelSignals(cancel)

	// 3. run rubik-agent
	if err := runAgent(ctx); err != nil {
		log.Errorf("error running rubik agent: %v", err)
		return constant.ErrorExitCode
	}
	return constant.NormalExitCode
}

func handelSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	for sig := range signalChan {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			log.Infof("signal %v received and starting exit...", sig)
			cancel()
		}
	}
}
