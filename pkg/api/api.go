package api

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Registry provides an interface for service discovery
type Registry interface {
	Init() error
	Register(*Service, string) error
	Deregister(*Service, string) error
	GetService(string) (*Service, error)
	ListServices() ([]*Service, error)
}

// Service contains progress that all services(modules) need to have
type Service interface {
	Init() error
	Setup() error
	Run() error
	TearDown() error
	PodEventHandler() error
}

// PodEventSubscriber control pod activities
type PodEventSubscriber interface {
	AddPod(pod *corev1.Pod)
	UpdatePod(pod *corev1.Pod)
	DeletePod(podID types.UID)
	ID() string
}

// Publisher publish pod event to all subscribers
type Publisher interface {
	Subscribe(s PodEventSubscriber)
	Unsubscribe(s PodEventSubscriber)
	NotifySubscribers()
	ReceivePodEvent(string, *corev1.Pod)
}

// Viewer collect on/offline pods info
type Viewer interface {
	ListOnlinePods() ([]*corev1.Pod, error)
	ListOfflinePods() ([]*corev1.Pod, error)
}
