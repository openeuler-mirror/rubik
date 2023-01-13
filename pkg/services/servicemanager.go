package services

import (
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

type ServiceManager struct {
	api.Subscriber
	sync.RWMutex
	RunningServices           map[string]api.Service
	RunningPersistentServices map[string]api.PersistentService
}

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

func (manager *ServiceManager) HandleEvent(eventType typedef.EventType, event typedef.Event) {
	switch eventType {
	case typedef.INFO_ADD:
	case typedef.INFO_UPDATE:
	case typedef.INFO_DELETE:
	default:
		log.Infof("service manager fail to process %s type", eventType.String())
	}
}

func (cmanager *ServiceManager) EventTypes() []typedef.EventType {
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
