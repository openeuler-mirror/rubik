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
// Description: This file defines ServiceManager to manage the lifecycle of services

// Package rubik implements service registration, discovery and management functions
package rubik

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/core/subscriber"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services"
)

// serviceManagerName is the unique ID of the service manager
const serviceManagerName = "serviceManager"

// ServiceManager is used to manage the lifecycle of services
type ServiceManager struct {
	api.Subscriber
	api.Viewer
	sync.RWMutex
	RunningServices map[string]services.Service
}

// NewServiceManager creates a servicemanager object
func NewServiceManager() *ServiceManager {
	manager := &ServiceManager{
		RunningServices: make(map[string]services.Service),
	}
	manager.Subscriber = subscriber.NewGenericSubscriber(manager, serviceManagerName)
	return manager
}

// InitServices parses the to-be-run services config and loads them to the ServiceManager
func (manager *ServiceManager) InitServices(features []string,
	serviceConfig map[string]interface{}, parser config.ConfigParser) error {
	for _, feature := range features {
		s, err := services.GetServiceComponent(feature)
		if err != nil {
			return fmt.Errorf("get component failed %s: %v", feature, err)
		}

		if err1 := s.SetConfig(func(configName string, v interface{}) error {
			config := serviceConfig[configName]
			if config == nil {
				return fmt.Errorf("this configuration is not available,configName:%v", configName)
			}
			if err2 := parser.UnmarshalSubConfig(config, v); err2 != nil {
				return fmt.Errorf("this configuration failed to be serialized,configName:%v,error:%v", configName, err2)
			}
			return nil
		}); err1 != nil {
			return fmt.Errorf("set configuration failed, err:%v", err1)
		}

		conf, err := parser.MarshalIndent(s.GetConfig(), "", "\t")
		if err != nil {
			return fmt.Errorf("failed to get service %v configuration: %v", s.ID(), err)
		}

		if err := manager.AddRunningService(feature, s); err != nil {
			return err
		}

		if len(conf) != 0 {
			log.Infof("service %v will run with configuration:\n%v", s.ID(), conf)
		} else {
			log.Infof("service %v will run", s.ID())
		}
	}
	return nil
}

// AddRunningService adds a to-be-run service
func (manager *ServiceManager) AddRunningService(name string, s services.Service) error {
	manager.Lock()
	defer manager.Unlock()
	if _, existed := manager.RunningServices[name]; existed {
		return fmt.Errorf("service name conflict: %s", name)
	}
	manager.RunningServices[name] = s
	return nil
}

// HandleEvent is used to handle PodInfo events pushed by the publisher
func (manager *ServiceManager) HandleEvent(eventType typedef.EventType, event typedef.Event) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic occur: %v", err)
		}
	}()
	switch eventType {
	case typedef.INFOADD:
		manager.addFunc(event)
	case typedef.INFOUPDATE:
		manager.updateFunc(event)
	case typedef.INFODELETE:
		manager.deleteFunc(event)
	default:
		log.Infof("service manager fail to process %s type", eventType.String())
	}
}

// EventTypes returns the type of event the serviceManager is interested in
func (manager *ServiceManager) EventTypes() []typedef.EventType {
	return []typedef.EventType{typedef.INFOADD, typedef.INFOUPDATE, typedef.INFODELETE}
}

// terminatingRunningServices handles services exits during the setup and exit phases
func terminatingServices(serviceMap map[string]services.Service, viewer api.Viewer) {
	for name, s := range serviceMap {
		if s.IsRunner() {
			if err := s.Stop(); err != nil {
				log.Errorf("failed to stop service %v: %v", name, err)
			} else {
				log.Infof("service %v stop successfully", name)
			}
		}
		if err := s.Terminate(viewer); err != nil {
			log.Errorf("failed to terminate service %v: %v", name, err)
		} else {
			log.Infof("service %v terminate successfully", name)
		}
	}
}

// Setup pre-starts services, such as preparing the environment, etc.
func (manager *ServiceManager) Setup(v api.Viewer) error {
	// only when viewer is prepared
	if v == nil {
		return fmt.Errorf("viewer should not be empty")
	}
	manager.Viewer = v

	var preStarted = make(map[string]services.Service, 0)
	manager.RLock()
	defer manager.RUnlock()
	for name, s := range manager.RunningServices {
		/*
			Try to prestart the service. If any service fails, rubik exits
			and invokes the terminate function to terminate the prestarted service.
		*/
		if err := s.PreStart(manager.Viewer); err != nil {
			terminatingServices(preStarted, manager.Viewer)
			return fmt.Errorf("failed to preStart service %v: %v", name, err)
		}
		preStarted[name] = s
		log.Infof("service %v pre-start successfully", name)
	}
	return nil
}

// Start starts and runs the persistent service
func (manager *ServiceManager) Start(ctx context.Context) {
	/*
		The Run function of the service will be called continuously until the context is canceled.
		When a service panics while running, recover will catch the violation
		and briefly restart for a short period of time.
	*/
	const restartDuration = 2 * time.Second
	runner := func(ctx context.Context, id string, runFunc func(ctx context.Context)) {
		var restartCount int64
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("service %s catch a panic: %v", id, err)
				}
			}()
			if restartCount > 0 {
				log.Warnf("service %s has restart %v times", id, restartCount)
			}
			restartCount++
			runFunc(ctx)
		}, restartDuration)
	}

	for id, s := range manager.RunningServices {
		if s.IsRunner() {
			go runner(ctx, id, s.Run)
		}
	}
}

// Stop terminates the running service
func (manager *ServiceManager) Stop() error {
	manager.RLock()
	terminatingServices(manager.RunningServices, manager.Viewer)
	manager.RUnlock()
	return nil
}

// addFunc handles pod addition events
func (manager *ServiceManager) addFunc(event typedef.Event) {
	podInfo, ok := event.(*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}

	const retryCount = 5
	addOnce := func(s services.Service, podInfo *typedef.PodInfo, wg *sync.WaitGroup) {
		for i := 0; i < retryCount; i++ {
			if err := s.AddPod(podInfo); err != nil {
				log.Errorf("service %s add func failed: %v", s.ID(), err)
			} else {
				break
			}
		}
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		wg.Add(1)
		go addOnce(s, podInfo.DeepCopy(), &wg)
	}
	wg.Wait()
	manager.RUnlock()
}

// updateFunc handles pod update events
func (manager *ServiceManager) updateFunc(event typedef.Event) {
	podInfos, ok := event.([]*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}
	const podInfosLen = 2
	if len(podInfos) != podInfosLen {
		log.Warnf("pod infos contains more than two pods")
		return
	}
	runOnce := func(s services.Service, old, new *typedef.PodInfo, wg *sync.WaitGroup) {
		log.Debugf("update Func with service: %s", s.ID())
		if err := s.UpdatePod(old, new); err != nil {
			log.Errorf("service %s update func failed: %v", s.ID(), err)
		}
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		wg.Add(1)
		go runOnce(s, podInfos[0], podInfos[1], &wg)
	}
	wg.Wait()
	manager.RUnlock()
}

// deleteFunc handles pod deletion events
func (manager *ServiceManager) deleteFunc(event typedef.Event) {
	podInfo, ok := event.(*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}

	deleteOnce := func(s services.Service, podInfo *typedef.PodInfo, wg *sync.WaitGroup) {
		if err := s.DeletePod(podInfo); err != nil {
			log.Errorf("service %s delete func failed: %v", s.ID(), err)
		}
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		wg.Add(1)
		go deleteOnce(s, podInfo.DeepCopy(), &wg)
	}
	wg.Wait()
	manager.RUnlock()
}
