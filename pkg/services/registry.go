package services

import (
	"sync"

	"isula.org/rubik/pkg/common/log"
)

type (
	// Creator creates Service objects
	Creator func() interface{}
	// registry is used for service registration
	registry struct {
		sync.RWMutex
		// services is a collection of all registered service
		services map[string]Creator
	}
)

// servicesRegistry  is the globally unique registry
var servicesRegistry = &registry{
	services: make(map[string]Creator, 0),
}

// Register is used to register the service creators
func Register(name string, creator Creator) {
	servicesRegistry.Lock()
	servicesRegistry.services[name] = creator
	servicesRegistry.Unlock()
	log.Debugf("func register (%s)", name)
}

// GetServiceCreator returns the service creator based on the incoming service name
func GetServiceCreator(name string) Creator {
	servicesRegistry.RLock()
	creator, ok := servicesRegistry.services[name]
	servicesRegistry.RUnlock()
	if !ok {
		return nil
	}
	return creator
}
