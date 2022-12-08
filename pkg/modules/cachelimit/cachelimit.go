package cachelimit

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/modules/checkpoint"
	"isula.org/rubik/pkg/registry"
)

type CacheLimit struct {
	Name string
}

func init() {
	registry.DefaultRegister.Register(NewCacheLimit(), "cacheLimit")
}

func NewCacheLimit() *CacheLimit {
	return &CacheLimit{Name: "cacheLimit"}
}

func (c *CacheLimit) Init() error {
	fmt.Println("cache limit Init()")
	return nil
}

func (c *CacheLimit) Setup() error {
	fmt.Println("cache limit Setup()")
	return nil
}

func (c *CacheLimit) Run() error {
	fmt.Println("cache limit Run()")
	return nil
}

func (c *CacheLimit) TearDown() error {
	fmt.Println("cache limit TearDown()")
	return nil
}

func (c *CacheLimit) AddPod(pod *corev1.Pod) {
	fmt.Println("cache limit AddPod()")
}

func (c *CacheLimit) UpdatePod(pod *corev1.Pod) {
	fmt.Println("cache limit UpdatePod()")
}

func (c *CacheLimit) DeletePod(podID types.UID) {
	fmt.Println("cache limit DeletePod()")
}

func (c *CacheLimit) ID() string {
	return c.Name
}

func (c *CacheLimit) PodEventHandler() error {
	fmt.Println("cache limit PodEventHandler")
	checkpoint.DefaultCheckPointPublisher.Subscribe(c)
	return nil
}
