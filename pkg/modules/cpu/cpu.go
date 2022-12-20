package cpu

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/modules/checkpoint"
	"isula.org/rubik/pkg/registry"
)

type CPU struct {
	Name string
}

func init() {
	registry.DefaultRegister.Register(NewCPU(), "cpu")
}

func NewCPU() *CPU {
	return &CPU{Name: "cpu"}
}

func (c *CPU) Init() error {
	fmt.Println("cpu Init()")
	return nil
}

func (c *CPU) Setup() error {
	fmt.Println("cpu Setup()")
	return nil
}

func (c *CPU) Run() error {
	fmt.Println("cpu Run()")
	return nil
}

func (c *CPU) TearDown() error {
	fmt.Println("cpu TearDown()")
	return nil
}

func (c *CPU) AddPod(pod *corev1.Pod) {
	fmt.Println("cpu AddPod()")
}

func (c *CPU) UpdatePod(pod *corev1.Pod) {
	fmt.Println("cpu UpdatePod()")
}

func (c *CPU) DeletePod(podID types.UID) {
	fmt.Println("cpu DeletePod()")
}

func (c *CPU) ID() string {
	return c.Name
}

func (c *CPU) PodEventHandler() error {
	fmt.Println("cpu PodEventHandler")
	checkpoint.DefaultCheckPointPublisher.Subscribe(c)
	return nil
}
