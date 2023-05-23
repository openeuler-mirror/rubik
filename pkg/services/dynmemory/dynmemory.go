package dynmemory

import (
	"context"
	"fmt"
	"time"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/services/helper"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DynMemoryAdapter is the adapter of dyn memory.
type DynMemoryAdapter interface {
	preStart(api.Viewer) error
	getInterval() int
	dynamicAdjust()
}
type dynMemoryConfig struct {
	Policy string `json:"policy,omitempty"`
}

// DynMemoryFactory is the factory of dyn memory.
type DynMemoryFactory struct {
	ObjName string
}

// Name to get the DynMemory factory name.
func (i DynMemoryFactory) Name() string {
	return "DynMemoryFactory"
}

// NewObj to create object of dynamic memory.
func (i DynMemoryFactory) NewObj() (interface{}, error) {
	return &DynMemory{ServiceBase: helper.ServiceBase{Name: i.ObjName}}, nil
}

// DynMemory is the struct of dynamic memory.
type DynMemory struct {
	helper.ServiceBase
	dynMemoryAdapter DynMemoryAdapter
}

// PreStart is an interface for calling a collection of methods when the service is pre-started
func (dynMem *DynMemory) PreStart(api api.Viewer) error {
	if dynMem.dynMemoryAdapter == nil {
		return nil
	}
	return dynMem.dynMemoryAdapter.preStart(api)
}

// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
func (dynMem *DynMemory) SetConfig(f helper.ConfigHandler) error {
	if f == nil {
		return fmt.Errorf("config handler function callback is nil")
	}

	var config dynMemoryConfig
	if err := f(dynMem.Name, &config); err != nil {
		return err
	}
	if dynMem.dynMemoryAdapter = newAdapter(config.Policy); dynMem.dynMemoryAdapter == nil {
		return fmt.Errorf("dynamic memory policy is error")
	}
	return nil
}

// Run implement service run function
func (dynMem *DynMemory) Run(ctx context.Context) {
	if dynMem.dynMemoryAdapter != nil {
		wait.Until(dynMem.dynMemoryAdapter.dynamicAdjust,
			time.Second*time.Duration(dynMem.dynMemoryAdapter.getInterval()),
			ctx.Done())
	} else {
		fmt.Println("dynamic memory policy is error")
	}
}

// IsRunner to Confirm whether it is a runner
func (dynMem *DynMemory) IsRunner() bool {
	return true
}

// newAdapter to create adapter of dyn memory.
func newAdapter(policy string) DynMemoryAdapter {
	switch policy {
	case "fssr":
		return initFssrDynMemAdapter()
	}
	return nil
}
