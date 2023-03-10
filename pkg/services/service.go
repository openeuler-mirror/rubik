package services

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services/helper"
)

// PodEvent for listening to pod changes.
type PodEvent interface {
	// Deal processing adding a pod.
	AddPod(podInfo *typedef.PodInfo) error
	// Deal processing update a pod config.
	UpdatePod(old, new *typedef.PodInfo) error
	// Deal processing delete a pod.
	DeletePod(podInfo *typedef.PodInfo) error
}

// Runner for background service process.
type Runner interface {
	// Confirm whether it is
	IsRunner() bool
	// Start runner
	Run(ctx context.Context)
	// Stop runner
	Stop() error
}

type HandlerConfig helper.HandlerConfig

// Service interface contains methods which must be implemented by all services.
type Service interface {
	Runner
	PodEvent
	// ID is the name of plugin, must be unique.
	ID() string
	// SetConfig is an interface that invoke the HandlerConfig to obtain the corresponding configuration.
	SetConfig(h HandlerConfig) error
	// PreStarter is an interface for calling a collection of methods when the service is pre-started
	PreStart(api.Viewer) error
	// Terminator is an interface that calls a collection of methods when the service terminates
	Terminate(api.Viewer) error
}

type FeatureSpec struct {
	// feature name
	Name string
	// Default is the default enablement state for the feature
	Default bool
}

func InitServiceComponents(specs []FeatureSpec) {
	for _, spec := range specs {
		if spec.Default {
			if initFunc, found := serviceComponents[spec.Name]; found {
				initFunc(spec.Name)
			} else {
				log.Errorf("init service failed, name:%v", spec.Name)
			}
		} else {
			log.Errorf("disable feature:%v", spec.Name)
		}
	}
}

func GetServiceComponent(name string) (Service, error) {
	if s, err := helper.GetComponent(name); err == nil {
		if si, ok := s.(Service); ok {
			return si, nil
		}
	}
	return nil, fmt.Errorf("get service failed, name:%v", name)
}
