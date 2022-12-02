package blkio

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/registry"
)

type Blkio struct {
	Name string
	api.Service
	api.PodEventSubscriber
}

func init() {
	registry.DefaultRegister.Register(&NewBlkio().Service, "blkio")
}

func NewBlkio() *Blkio {
	return &Blkio{Name: "blkio"}
}

func (b *Blkio) Init() error {
	fmt.Println("blkio Init()")
	return nil
}

func (b *Blkio) Setup() error {
	fmt.Println("blkio Setup()")
	return nil
}

func (b *Blkio) Run() error {
	fmt.Println("blkio Run()")
	return nil
}

func (b *Blkio) TearDown() error {
	fmt.Println("blkio TearDown()")
	return nil
}

func (b *Blkio) AddPod(pod *corev1.Pod) {
	fmt.Println("blkio AddPod()")
}

func (b *Blkio) UpdatePod(pod *corev1.Pod) {
	fmt.Println("blkio UpdatePod()")
}

func (b *Blkio) DeletePod(podID types.UID) {
	fmt.Println("blkio DeletePod()")
}

func (b *Blkio) ID() string {
	return b.Name
}
