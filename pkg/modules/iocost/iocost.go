package iocost

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/modules/checkpoint"
	"isula.org/rubik/pkg/registry"
)

type IOCost struct {
	Name string
}

func init() {
	registry.DefaultRegister.Register(NewIOCost(), "iocost")
}

func NewIOCost() *IOCost {
	return &IOCost{Name: "iocost"}
}

func (i *IOCost) Init() error {
	fmt.Println("iocost Init()")
	return nil
}

func (i *IOCost) Setup() error {
	fmt.Println("iocost Setup()")
	return nil
}

func (i *IOCost) Run() error {
	fmt.Println("iocost Run()")
	return nil
}

func (i *IOCost) TearDown() error {
	fmt.Println("iocost TearDown()")
	return nil
}

func (i *IOCost) AddPod(pod *corev1.Pod) {
	fmt.Println("iocost AddPod()")
}

func (i *IOCost) UpdatePod(pod *corev1.Pod) {
	fmt.Println("iocost UpdatePod()")
}

func (i *IOCost) DeletePod(podID types.UID) {
	fmt.Println("iocost DeletePod()")
}

func (i *IOCost) ID() string {
	return i.Name
}

func (i *IOCost) PodEventHandler() error {
	fmt.Println("iosoct PodEventHandler")
	checkpoint.DefaultCheckPointPublisher.Subscribe(i)
	return nil
}
