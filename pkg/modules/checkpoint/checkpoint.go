package checkpoint

import (
	"fmt"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/typedef"
	corev1 "k8s.io/api/core/v1"
)

type CheckPoint struct {
	Pods map[string]*typedef.PodInfo `json:"pods,omitempty"`
}

type CheckPointViewer struct {
	checkpoint  CheckPoint
	subscribers []api.PodEventSubscriber
}

type CheckPointPublisher struct {
	checkpoint  *CheckPoint
	subscribers []api.PodEventSubscriber
	viewer      CheckPointViewer
}

var DefaultCheckPointPublisher = NewCheckPointPublisher()

func NewCheckPointPublisher() *CheckPointPublisher {
	return &CheckPointPublisher{
		checkpoint: &CheckPoint{
			Pods: make(map[string]*typedef.PodInfo),
		},
		subscribers: make([]api.PodEventSubscriber, 0),
	}
}

func (cpp *CheckPointPublisher) Subscribe(s api.PodEventSubscriber) {
	fmt.Printf("CheckPointPublisher subscribe(%s)\n", s.ID())
	cpp.subscribers = append(cpp.subscribers, s)
}

func (cpp *CheckPointPublisher) Unsubscribe(s api.PodEventSubscriber) {
	fmt.Printf("CheckPointPublisher unsubscribe(%s)\n", s.ID())
	subscribersLength := len(cpp.subscribers)
	for i, subscriber := range cpp.subscribers {
		if s.ID() == subscriber.ID() {
			cpp.subscribers[subscribersLength-1], cpp.subscribers[i] = cpp.subscribers[i], cpp.subscribers[subscribersLength-1]
			cpp.subscribers = cpp.subscribers[:subscribersLength-1]
			break
		}
	}
}

func (cpp *CheckPointPublisher) NotifySubscribers() {
	fmt.Printf("CheckPointPublisher notifyAll()\n")
}

func (cpp *CheckPointPublisher) ReceivePodEvent(eventType string, pod *corev1.Pod) {
	fmt.Printf("CheckPointPublisher ReceivePodEvent(%s)\n", eventType)
}

func (cv *CheckPointViewer) ListOnlinePods() ([]*corev1.Pod, error) {
	fmt.Printf("CheckPointViewer ListOnlinePods()\n")
	return nil, nil
}

func (cv *CheckPointViewer) ListOfflinePods() ([]*corev1.Pod, error) {
	fmt.Printf("CheckPointViewer ListOfflinePods()\n")
	return nil, nil
}
