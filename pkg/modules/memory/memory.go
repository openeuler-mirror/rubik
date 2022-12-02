package memory

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/registry"
)

type Memory struct {
	Name string
	api.Service
	api.PodEventSubscriber
}

func init() {
	registry.DefaultRegister.Register(&NewMemory().Service, "memory")
}

func NewMemory() *Memory {
	return &Memory{Name: "memory"}
}

func (m *Memory) Init() error {
	fmt.Println("memory Init()")
	return nil
}

func (m *Memory) Setup() error {
	fmt.Println("memory Setup()")
	return nil
}

func (m *Memory) Run() error {
	fmt.Println("memory Run()")
	return nil
}

func (m *Memory) TearDown() error {
	fmt.Println("memory TearDown()")
	return nil
}

func (m *Memory) AddPod(pod *corev1.Pod) {
	fmt.Println("memory AddPod()")
}

func (m *Memory) UpdatePod(pod *corev1.Pod) {
	fmt.Println("memory UpdatePod()")
}

func (m *Memory) DeletePod(podID types.UID) {
	fmt.Println("memory DeletePod()")
}

func (m *Memory) ID() string {
	return m.Name
}
