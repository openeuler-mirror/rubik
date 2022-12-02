package registry

import (
	"fmt"

	"isula.org/rubik/pkg/api"
)

type RubikRegistry struct {
	services map[string]*api.Service
}

var DefaultRegister = NewRegistry()

func NewRegistry() *RubikRegistry {
	return &RubikRegistry{
		services: make(map[string]*api.Service),
	}
}

func (r *RubikRegistry) Init() error {
	fmt.Println("rubik registry Init()")
	return nil
}

func (r *RubikRegistry) Register(s *api.Service, name string) error {
	fmt.Printf("rubik registry Register(%s)\n", name)
	if _, ok := r.services[name]; !ok {
		r.services[name] = s
	}
	return nil
}

func (r *RubikRegistry) Deregister(s *api.Service, name string) error {
	fmt.Printf("rubik register Deregister(%s)\n", name)
	delete(r.services, name)
	return nil
}

func (r *RubikRegistry) GetService(name string) (*api.Service, error) {
	fmt.Printf("rubik register GetService(%s)\n", name)
	if s, ok := r.services[name]; ok {
		return s, nil
	} else {
		return nil, fmt.Errorf("service %s did not registered", name)
	}
}

func (r *RubikRegistry) ListServices() ([]*api.Service, error) {
	fmt.Println("rubik register ListServices()")
	var services []*api.Service
	for _, s := range r.services {
		services = append(services, s)
	}
	return services, nil
}
