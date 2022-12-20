package rubik

import (
	"fmt"

	"isula.org/rubik/pkg/registry"
)

func Run() int {
	fmt.Println("rubik running")

	// 0. services automatic registration
	// done in import.go

	// 1. enable autoconfig(informer)
	// podinformer.Init()

	// 2. enable checkpoint
	services, err := registry.DefaultRegister.ListServices()
	if err != nil {
		return -1
	}
	for _, s := range services {
		if err := s.PodEventHandler(); err != nil {
			continue
		}
		s.Init()
		s.Setup()
		s.Run()
		defer s.TearDown()
	}

	return 0
}
