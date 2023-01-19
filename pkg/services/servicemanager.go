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
	api.Viewer
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
		manager.AddFunc(event)
	case typedef.INFO_UPDATE:
		manager.UpdateFunc(event)
	case typedef.INFO_DELETE:
		manager.DeleteFunc(event)
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

func (manager *ServiceManager) Setup() error {
	for _, s := range manager.RunningServices {
		if err := s.Setup(); err != nil {
			s.TearDown()
			return err
		}
	}
	for _, s := range manager.RunningPersistentServices {
		if err := s.Setup(manager.Viewer); err != nil {
			return err
		}
	}
	return nil
}

func (manager *ServiceManager) Run(stopChan chan struct{}) error {
	for _, s := range manager.RunningPersistentServices {
		s.Start(stopChan)
	}
	return nil
}

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
	manager.RUnlock()
}

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
	manager.RUnlock()
}

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
	manager.RUnlock()
}
