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
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/core/subscriber"
	"isula.org/rubik/pkg/core/typedef"
)

// serviceManagerName is the unique ID of the service manager
const serviceManagerName = "serviceManager"

// ServiceManager is used to manage the lifecycle of services
type ServiceManager struct {
	api.Subscriber
	api.Viewer
	sync.RWMutex
	RunningServices           map[string]api.Service
	RunningPersistentServices map[string]api.PersistentService
	TerminateFuncs            map[string]Terminator
}

// NewServiceManager creates a servicemanager object
func NewServiceManager() *ServiceManager {
	manager := &ServiceManager{
		RunningServices:           make(map[string]api.Service),
		RunningPersistentServices: make(map[string]api.PersistentService),
	}
	manager.Subscriber = subscriber.NewGenericSubscriber(manager, serviceManagerName)
	return manager
}

// InitServices parses the to-be-run services config and loads them to the ServiceManager
func (manager *ServiceManager) InitServices(serviceConfig map[string]interface{}, parser config.ConfigParser) error {
	for name, config := range serviceConfig {
		creator := GetServiceCreator(name)
		if creator == nil {
			return fmt.Errorf("no corresponding module %v, please check the module name", name)
		}
		service := creator()
		if err := parser.UnmarshalSubConfig(config, service); err != nil {
			return fmt.Errorf("error unmarshaling %s config: %v", name, err)
		}

		if SetLoggerOnService(service,
			log.WithCtx(context.WithValue(context.Background(), log.CtxKey(constant.LogEntryKey), name))) {
			log.Debugf("set logger for service: %s", name)
		}

		// try to verify configuration
		if validator, ok := service.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("error configuring service %s: %v", name, err)
			}
		}
		if err := manager.AddRunningService(name, service); err != nil {
			return err
		}
	}
	return nil
}

// AddRunningService adds a to-be-run service
func (manager *ServiceManager) AddRunningService(name string, service interface{}) error {
	manager.RLock()
	_, existed1 := manager.RunningServices[name]
	_, existed2 := manager.RunningPersistentServices[name]
	manager.RUnlock()
	if existed1 || existed2 {
		return fmt.Errorf("service name conflict: %s", name)
	}

	addService := manager.tryAddService(name, service)
	addPersistentService := manager.tryAddPersistentService(name, service)
	if addPersistentService || addService {
		log.Debugf("pre-start service %s", name)
		return nil
	}
	return fmt.Errorf("invalid service %s (type %T)", name, service)
}

// HandleEvent is used to handle PodInfo events pushed by the publisher
func (manager *ServiceManager) HandleEvent(eventType typedef.EventType, event typedef.Event) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic occurr: %v", err)
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

// tryAddService determines whether it is a api.Service and adds it to the queue to be run
func (manager *ServiceManager) tryAddService(name string, service interface{}) bool {
	s, ok := service.(api.Service)
	if ok {
		manager.Lock()
		manager.RunningServices[name] = s
		manager.Unlock()
		log.Debugf("service %s will run", name)
	}
	return ok
}

// tryAddPersistentService determines whether it is a api.PersistentService and adds it to the queue to be run
func (manager *ServiceManager) tryAddPersistentService(name string, service interface{}) bool {
	s, ok := service.(api.PersistentService)
	if ok {
		manager.Lock()
		manager.RunningPersistentServices[name] = s
		manager.Unlock()
		log.Debugf("persistent service %s will run", name)
	}
	return ok
}

// terminatingRunningServices handles services exits during the setup and exit phases
func (manager *ServiceManager) terminatingRunningServices(err error) error {
	if manager.TerminateFuncs == nil {
		return nil
	}
	for id, t := range manager.TerminateFuncs {
		if termErr := t.Terminate(manager.Viewer); termErr != nil {
			log.Errorf("error terminating services %s: %v", id, termErr)
		}
	}
	return err
}

// SetLoggerOnService assigns a value to the variable Log member if there is a Log field
func SetLoggerOnService(value interface{}, logger api.Logger) bool {
	// 1. call the SetupLog function to set up the log
	method := reflect.ValueOf(value).MethodByName("SetupLog")
	if method.IsValid() && !method.IsZero() && !method.IsNil() {
		method.Call([]reflect.Value{reflect.ValueOf(logger)})
		return true
	}

	// 2. look for a member variable named Log
	field := reflect.ValueOf(value).Elem().FieldByName("Log")
	if field.IsValid() && field.CanSet() && field.Type().String() == "api.Logger" {
		field.Set(reflect.ValueOf(logger))
		return true
	}

	return false
}

// Setup pre-starts services, such as preparing the environment, etc.
func (manager *ServiceManager) Setup(v api.Viewer) error {
	// only when viewer is prepared
	if v == nil {
		return nil
	}
	preStarted := make(map[string]struct{}, 0)
	manager.Viewer = v
	manager.TerminateFuncs = make(map[string]Terminator)
	setupFunc := func(id string, s interface{}) error {
		// 1. record the termination function of the service that has been setup
		if t, ok := s.(Terminator); ok {
			manager.TerminateFuncs[id] = t
		}
		// 2. execute the pre-start function of the service
		p, ok := s.(PreStarter)
		if !ok {
			return nil
		}
		// execute the prestart method only once if the service is both a persistent service and a service
		if _, ok := preStarted[id]; ok {
			return nil
		}
		if err := p.PreStart(manager.Viewer); err != nil {
			return err
		}
		preStarted[id] = struct{}{}
		return nil
	}

	// 1. pre-start services
	for _, s := range manager.RunningServices {
		if err := setupFunc(s.ID(), s); err != nil {
			/*
				handle the error and terminate all services that have been started
				when any setup stage failed
			*/
			return manager.terminatingRunningServices(fmt.Errorf("error running services %s: %v", s.ID(), err))
		}
	}
	// 2. pre-start persistent services
	for _, s := range manager.RunningPersistentServices {
		if err := setupFunc(s.ID(), s); err != nil {
			return manager.terminatingRunningServices(fmt.Errorf("error running persistent services %s: %v", s.ID(), err))
		}
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
		var restartCount int64 = 0
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
	for id, s := range manager.RunningPersistentServices {
		go runner(ctx, id, s.Run)
	}
}

// Stop terminates the running service
func (manager *ServiceManager) Stop() error {
	return manager.terminatingRunningServices(nil)
}

// addFunc handles pod addition events
func (manager *ServiceManager) addFunc(event typedef.Event) {
	podInfo, ok := event.(*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}

	const retryCount = 5
	runOnce := func(s api.Service, podInfo *typedef.PodInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		for i := 0; i < retryCount; i++ {
			if err := s.AddFunc(podInfo); err != nil {
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
		go runOnce(s, podInfo.DeepCopy(), &wg)
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
	runOnce := func(s api.Service, old, new *typedef.PodInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		log.Debugf("update Func with service: %s", s.ID())
		if err := s.UpdateFunc(old, new); err != nil {
			log.Errorf("service %s update func failed: %v", s.ID(), err)
		}
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

// deleteFunc handles pod deletion events
func (manager *ServiceManager) deleteFunc(event typedef.Event) {
	podInfo, ok := event.(*typedef.PodInfo)
	if !ok {
		log.Warnf("receive invalid event: %T", event)
		return
	}

	runOnce := func(s api.Service, podInfo *typedef.PodInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		if err := s.DeleteFunc(podInfo); err != nil {
			log.Errorf("service %s delete func failed: %v", s.ID(), err)
		}
		wg.Done()
	}
	manager.RLock()
	var wg sync.WaitGroup
	for _, s := range manager.RunningServices {
		go runOnce(s, podInfo.DeepCopy(), &wg)
	}
	wg.Wait()
	manager.RUnlock()
}

// Terminator is an interface that calls a collection of methods when the service terminates
type Terminator interface {
	Terminate(api.Viewer) error
}

// PreStarter is an interface for calling a collection of methods when the service is pre-started
type PreStarter interface {
	PreStart(api.Viewer) error
}

// Validator is a function interface to verify whether the configuration is correct or not
type Validator interface {
	Validate() error
}
