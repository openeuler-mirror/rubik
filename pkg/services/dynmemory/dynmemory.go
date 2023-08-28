package dynmemory

import (
	"context"
	"fmt"
	"time"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services/helper"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DynMemoryAdapter is the adapter of dyn memory.
type DynMemoryAdapter interface {
	preStart(api.Viewer) error
	getInterval() int
	dynamicAdjust()
	setOfflinePod(path string) error
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
func (dynMem *DynMemory) PreStart(viewer api.Viewer) error {
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	if dynMem.dynMemoryAdapter == nil {
		return nil
	}
	return dynMem.dynMemoryAdapter.preStart(viewer)
}

// SetConfig is an interface that invoke the ConfigHandler to obtain the corresponding configuration.
func (dynMem *DynMemory) SetConfig(f helper.ConfigHandler) error {
	if f == nil {
		return fmt.Errorf("no config handler function callback")
	}

	var config dynMemoryConfig
	if err := f(dynMem.Name, &config); err != nil {
		return err
	}
	if dynMem.dynMemoryAdapter = newAdapter(config.Policy); dynMem.dynMemoryAdapter == nil {
		return fmt.Errorf("invalid dynamic memory policy")
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
		fmt.Println("invalid dynamic memory policy")
	}
}

// IsRunner to Confirm whether it is a runner
func (dynMem *DynMemory) IsRunner() bool {
	return true
}

// AddPod to deal the event of adding a pod.
func (dynMem *DynMemory) AddPod(podInfo *typedef.PodInfo) error {
	if podInfo.Offline() {
		return dynMem.dynMemoryAdapter.setOfflinePod(podInfo.Path)
	}
	return nil
}

// UpdatePod to deal the pod update event.
func (dynMem *DynMemory) UpdatePod(old, new *typedef.PodInfo) error {
	if new.Offline() {
		return dynMem.dynMemoryAdapter.setOfflinePod(new.Path)
	}
	return nil
}

// newAdapter to create adapter of dyn memory.
func newAdapter(policy string) DynMemoryAdapter {
	switch policy {
	case "fssr":
		return initFssrDynMemAdapter()
	default:
		log.Errorf("no matching policy[%v] is found", policy)
	}
	return nil
}
