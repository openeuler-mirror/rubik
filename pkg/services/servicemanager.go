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

// Package services implements service registration, discovery and management functions
package services

import (
	"fmt"
	"sync"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/subscriber"
	"isula.org/rubik/pkg/core/typedef"
)

// serviceManagerName is the unique ID of the service manager
const serviceManagerName = "serviceManager"

// AddRunningService adds a to-be-run service
func AddRunningService(name string, service interface{}) {
	serviceManager.RLock()
	_, ok1 := serviceManager.RunningServices[name]
	_, ok2 := serviceManager.RunningPersistentServices[name]
	serviceManager.RUnlock()
	if ok1 || ok2 {
		log.Errorf("service name conflict : \"%s\"", name)
		return
	}

	if !serviceManager.tryAddService(name, service) && !serviceManager.tryAddPersistentService(name, service) {
		log.Errorf("invalid service : \"%s\", %T", name, service)
		return
	}
	log.Debugf("pre-start service %s", name)
}

// ServiceManager is used to manage the lifecycle of services
type ServiceManager struct {
	api.Subscriber
	api.Viewer
	sync.RWMutex
	RunningServices           map[string]api.Service
	RunningPersistentServices map[string]api.PersistentService
	exitFuncs                 []func() error
}

// serviceManager is the only global service manager
var serviceManager = newServiceManager()

func newServiceManager() *ServiceManager {
	manager := &ServiceManager{
		RunningServices:           make(map[string]api.Service),
		RunningPersistentServices: make(map[string]api.PersistentService),
	}
	manager.Subscriber = subscriber.NewGenericSubscriber(manager, serviceManagerName)
	return manager
}

// GetServiceManager returns the globally unique ServiceManager
func GetServiceManager() *ServiceManager {
	return serviceManager
}

// HandleEvent is used to handle PodInfo events pushed by the publisher
func (manager *ServiceManager) HandleEvent(eventType typedef.EventType, event typedef.Event) {
	switch eventType {
	case typedef.INFO_ADD:
		manager.AddFunc(event)
	case typedef.INFO_UPDATE:
		manager.UpdateFunc(event)
	case typedef.INFO_DELETE:
		manager.DeleteFunc(event)
	default:
		log.Infof("service manager fail to process %s type", eventType.String())
	}
}

// EventTypes returns the type of event the serviceManager is interested in
func (manager *ServiceManager) EventTypes() []typedef.EventType {
	return []typedef.EventType{typedef.INFO_ADD, typedef.INFO_UPDATE, typedef.INFO_DELETE}
}

// tryAddService determines whether it is a api.Service and adds it to the queue to be run
func (manager *ServiceManager) tryAddService(name string, service interface{}) bool {
	s, ok := service.(api.Service)
	if ok {
		serviceManager.Lock()
		manager.RunningServices[name] = s
		serviceManager.Unlock()
		log.Debugf("service %s will run", name)
	}
	return ok
}

// tryAddPersistentService determines whether it is a api.PersistentService and adds it to the queue to be run
func (manager *ServiceManager) tryAddPersistentService(name string, service interface{}) bool {
	s, ok := service.(api.PersistentService)
	if ok {
		serviceManager.Lock()
		manager.RunningPersistentServices[name] = s
		serviceManager.Unlock()
		log.Debugf("persistent service %s will run", name)
	}
	return ok
}

// Setup pre-starts services, such as preparing the environment, etc.
func (manager *ServiceManager) Setup(v api.Viewer) error {
	var (
		exitFuncs []func() error
		// process when fail to pre-start services
		errorHandler = func(err error) error {
			for _, exitFunc := range exitFuncs {
				exitFunc()
			}
			return err
		}
	)
	// pre-start services
	for _, s := range manager.RunningServices {
		if err := s.PreStart(); err != nil {
			return errorHandler(fmt.Errorf("error running services %s: %v", s.ID(), err))
		}
		if t, ok := s.(Terminator); ok {
			exitFuncs = append(exitFuncs, t.Terminate)
		}
	}
	// pre-start persistent services only when viewer is prepared
	if v == nil {
		return nil
	}
	manager.Viewer = v
	for _, s := range manager.RunningPersistentServices {
		if err := s.PreStart(manager.Viewer); err != nil {
			return errorHandler(fmt.Errorf("error running persistent services %s: %v", s.ID(), err))
		}
		if t, ok := s.(Terminator); ok {
			exitFuncs = append(exitFuncs, t.Terminate)
		}
	}
	manager.exitFuncs = exitFuncs
	return nil
}

// Start starts and runs the service
func (manager *ServiceManager) Start(stopChan <-chan struct{}) error {
	for _, s := range manager.RunningPersistentServices {
		s.Start(stopChan)
	}
	return nil
}

// Start starts and runs the service
func (manager *ServiceManager) Stop() error {
	var (
		allErr     error
		errorOccur bool = false
	)
	for _, exitFunc := range manager.exitFuncs {
		if err := exitFunc(); err != nil {
			log.Errorf("error stopping services: %v", err)
			allErr = fmt.Errorf("error stopping services")
			errorOccur = true
		}
	}
	if errorOccur {
		return allErr
	}
	return nil
}

// AddFunc handles pod addition events
func (manager *ServiceManager) AddFunc(event typedef.Event) {
	podInfo, ok := event.(*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}

	runOnce := func(s api.Service, podInfo *typedef.PodInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		s.AddFunc(podInfo)
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		go runOnce(s, podInfo.Clone(), &wg)
	}
	wg.Wait()
	manager.RUnlock()
}

// UpdateFunc() handles pod update events
func (manager *ServiceManager) UpdateFunc(event typedef.Event) {
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
	runOnce := func(s api.Service, old, new *typedef.PodInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		s.UpdateFunc(old, new)
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		go runOnce(s, podInfos[0], podInfos[1], &wg)
	}
	wg.Wait()
	manager.RUnlock()
}

// DeleteFunc handles pod deletion events
func (manager *ServiceManager) DeleteFunc(event typedef.Event) {
	podInfo, ok := event.(*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}

	runOnce := func(s api.Service, podInfo *typedef.PodInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		s.DeleteFunc(podInfo)
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		go runOnce(s, podInfo.Clone(), &wg)
	}
	wg.Wait()
	manager.RUnlock()
}

type Terminator interface {
	Terminate() error
}
