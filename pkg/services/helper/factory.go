package helper

import (
	"errors"
	"sync"
)

type ServiceFactory interface {
	Name() string
	NewObj() (interface{}, error)
}

var (
	rwlock           sync.RWMutex
	serviceFactories = map[string]ServiceFactory{}
)

func AddFactory(name string, factory ServiceFactory) {
	rwlock.Lock()
	defer rwlock.Unlock()
	serviceFactories[name] = factory
}

func GetComponent(name string) (interface{}, error) {
	rwlock.RLock()
	defer rwlock.RUnlock()
	if f, found := serviceFactories[name]; found {
		return f.NewObj()
	} else {
		return nil, errors.New("factory is not found")
	}
}
