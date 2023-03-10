package helper

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
)

type ServiceBase struct{}
type HandlerConfig func(configName string, d interface{}) error

// ID is the name of plugin, must be unique.
func (s *ServiceBase) ID() string {
	panic("this interface must be implemented.")
}

// PreStarter is an interface for calling a collection of methods when the service is pre-started
func (s *ServiceBase) PreStart(api.Viewer) error {
	return nil
}

// Terminator is an interface that calls a collection of methods when the service terminates
func (s *ServiceBase) Terminate(api.Viewer) error {
	return nil
}

// Confirm whether it is
func (s *ServiceBase) IsRunner() bool {
	return false
}

// Start runner
func (s *ServiceBase) Run(ctx context.Context) {}

// Stop runner
func (s *ServiceBase) Stop() error {
	return fmt.Errorf("i am not runner")
}

func (s *ServiceBase) AddPod(podInfo *typedef.PodInfo) error {
	return nil
}

func (S *ServiceBase) UpdatePod(old, new *typedef.PodInfo) error {
	return nil
}

func (s *ServiceBase) DeletePod(podInfo *typedef.PodInfo) error {
	return nil
}

func (s *ServiceBase) SetConfig(h HandlerConfig) error {
	return nil
}
