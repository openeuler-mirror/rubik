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
	"isula.org/rubik/pkg/core/trigger"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/informer"
	"isula.org/rubik/pkg/podmanager"
	"isula.org/rubik/pkg/services"
)

// Agent runs a series of rubik services and manages data
type Agent struct {
	config          *config.Config
	podManager      *podmanager.PodManager
	informer        api.Informer
	servicesManager *ServiceManager
}

// NewAgent returns an agent for given configuration
func NewAgent(cfg *config.Config) (*Agent, error) {
	publisher := publisher.GetPublisherFactory().GetPublisher(publisher.GENERIC)
	serviceManager := NewServiceManager()
	if err := serviceManager.InitServices(cfg.Agent.EnabledFeatures,
		cfg.UnwrapServiceConfig(), cfg); err != nil {
		return nil, err
	}
	a := &Agent{
		config:          cfg,
		podManager:      podmanager.NewPodManager(publisher),
		servicesManager: serviceManager,
	}
	return a, nil
}

// Run starts and runs the agent until receiving stop signal
func (a *Agent) Run(ctx context.Context) error {
	log.Infof("agent run with config:\n%s", a.config.String())
	var informerName string
	informerName = a.config.Agent.InformerType
	if informerName == "nil" {
		informerName = constant.APIServerInformer
	}
	if err := a.startInformer(ctx, informerName); err != nil {
		return err
	}
	a.informer.WaitReady()
	if err := a.startServiceHandler(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	a.stopInformer()
	a.stopServiceHandler()
	return nil
}

// startInformer starts informer to obtain external data
func (a *Agent) startInformer(ctx context.Context, informerName string) error {
	publisher := publisher.GetPublisherFactory().GetPublisher(publisher.GENERIC)
	var i api.Informer
	var err error
	if informerName == constant.APIServerInformer {
		i, err = informer.GetInformerFactory().GetInformerCreator(informer.APISERVER)(publisher)
	} else {
		i, err = informer.GetInformerFactory().GetInformerCreator(informer.NRI)(publisher)
	}
	if err != nil {
		return fmt.Errorf("failed to set informer: %v", err)
	}
	if err := i.Subscribe(a.podManager); err != nil {
		return fmt.Errorf("failed to subscribe informer: %v", err)
	}
	a.informer = i
	i.Start(ctx)
	return nil
}

// stopInformer stops the informer
func (a *Agent) stopInformer() {
	a.informer.Unsubscribe(a.podManager)
}

// startServiceHandler starts and runs the service
func (a *Agent) startServiceHandler(ctx context.Context) error {
	if err := a.servicesManager.Setup(a.podManager); err != nil {
		return fmt.Errorf("failed to set service handler: %v", err)
	}
	a.servicesManager.Start(ctx)
	a.podManager.Subscribe(a.servicesManager)
	return nil
}

// stopServiceHandler stops sending data to the ServiceManager
func (a *Agent) stopServiceHandler() {
	a.podManager.Unsubscribe(a.servicesManager)
	a.servicesManager.Stop()
}

// runAgent creates and runs rubik's agent
func runAgent(ctx context.Context) error {
	// 1. read configuration
	c := config.NewConfig(config.JSON)
	if err := c.LoadConfig(constant.ConfigFile); err != nil {
		return fmt.Errorf("failed to load agent config: %v", err)
	}

	// 2. enable log system
	if err := log.InitConfig(c.Agent.LogDriver, c.Agent.LogDir, c.Agent.LogLevel, c.Agent.LogSize); err != nil {
		return fmt.Errorf("failed to initialize log: %v", err)
	}

	// 3. enable cgroup system
	cgroup.InitMountDir(c.Agent.CgroupRoot)
	if c.Agent.CgroupDriver != "" {
		cgroup.SetCgroupDriver(c.Agent.CgroupDriver)
	}

	// 4. init service components
	services.InitServiceComponents(defaultRubikFeature)

	// 5. Create and run the agent
	agent, err := NewAgent(c)
	if err != nil {
		return fmt.Errorf("failed to create agent: %v", err)
	}
	if err := agent.Run(ctx); err != nil {
		return fmt.Errorf("failed to start agent: %v", err)
	}
	return nil
}

// Run runs agent and process signal
func Run() int {
	// 0. file mask permission setting and parameter checking
	unix.Umask(constant.DefaultUmask)
	if len(os.Args) > 1 {
		fmt.Println("invalid parameter, the length must be greater than 1")
		return constant.ArgumentErrorExitCode
	}
	// 1. apply file locks, only one rubik process can run at the same time
	lock, err := util.LockFile(constant.LockFile)
	defer func() {
		lock.Close()
		log.DropError(os.Remove(constant.LockFile))
	}()
	if err != nil {
		fmt.Printf("set rubik lock failed: %v, check if there is another rubik running\n", err)
		return constant.RepeatRunExitCode
	}
	defer util.UnlockFile(lock)

	ctx, cancel := context.WithCancel(context.Background())
	// 2. handle external signals
	go handleSignals(cancel)

	// 3. run rubik-agent
	defer cleanDepecdencies()
	if err := runAgent(ctx); err != nil {
		log.Errorf("failed to run rubik agent: %v", err)
		return constant.ErrorExitCode
	}
	return constant.NormalExitCode
}

func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	for sig := range signalChan {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			log.Infof("signal %v received and starting exit...", sig)
			cancel()
		}
	}
}

func cleanDepecdencies() {
	if err := trigger.StopUsedExecutors(); err != nil {
		log.Errorf("failed to stop dependencies: %v", err)
	}
}
