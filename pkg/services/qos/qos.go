package qos

import (
	"fmt"
	"strconv"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services"
)

var supportCgroupTypes = map[string]*cgroup.Key{
	"cpu":    {SubSys: "cpu", FileName: constant.CPUCgroupFileName},
	"memory": {SubSys: "memory", FileName: constant.MemoryCgroupFileName},
}

// QoS define service which related to qos level setting
type QoS struct {
	Name string `json:"-"`
	Log  api.Logger
	Config
}

// Config contains sub-system that need to set qos level
type Config struct {
	SubSys []string `json:"subSys"`
}

func init() {
	services.Register("qos", func() interface{} {
		return NewQoS()
	})
}

// NewQoS return qos instance
func NewQoS() *QoS {
	return &QoS{
		Name: "qos",
	}
}

// ID return qos service name
func (q *QoS) ID() string {
	return q.Name
}

// AddFunc implement add function when pod is added in k8s
func (q *QoS) AddFunc(pod *typedef.PodInfo) error {
	if err := q.SetQoS(pod); err != nil {
		return err
	}
	if err := q.ValidateQoS(pod); err != nil {
		return err
	}
	return nil
}

// UpdateFunc implement update function when pod info is changed
func (q *QoS) UpdateFunc(old, new *typedef.PodInfo) error {
	oldQos, newQos := getQoSLevel(old), getQoSLevel(new)
	switch {
	case newQos == oldQos:
		return nil
	case newQos > oldQos:
		return fmt.Errorf("not support change qos level from low to high")
	default:
		if err := q.ValidateQoS(new); err != nil {
			if err := q.SetQoS(new); err != nil {
				return fmt.Errorf("update qos for pod %s(%s) failed: %v", new.Name, new.UID, err)
			}
		}
	}
	return nil
}

// DeleteFunc implement delete function when pod is deleted by k8s
func (q *QoS) DeleteFunc(pod *typedef.PodInfo) error {
	return nil
}

// ValidateQoS will validate pod's qos level between value from
// cgroup file and the one from pod info
func (q *QoS) ValidateQoS(pod *typedef.PodInfo) error {
	targetLevel := getQoSLevel(pod)
	for _, subSys := range q.SubSys {
		if err := pod.GetCgroupAttr(supportCgroupTypes[subSys]).Expect(targetLevel); err != nil {
			return fmt.Errorf("failed to validate pod %s: %v", pod.Name, err)
		}
		for _, container := range pod.IDContainersMap {
			if err := container.GetCgroupAttr(supportCgroupTypes[subSys]).Expect(targetLevel); err != nil {
				return fmt.Errorf("failed to validate pod %s: %v", pod.Name, err)
			}
		}
	}
	return nil
}

// SetQoS set pod and all containers' qos level within it
func (q *QoS) SetQoS(pod *typedef.PodInfo) error {
	qosLevel := getQoSLevel(pod)
	if qosLevel != constant.Offline {
		q.Log.Debugf("pod %s already online", pod.Name)
		return nil
	}

	for _, sys := range q.SubSys {
		if err := pod.SetCgroupAttr(supportCgroupTypes[sys], strconv.Itoa(qosLevel)); err != nil {
			return err
		}
		for _, container := range pod.IDContainersMap {
			if err := container.SetCgroupAttr(supportCgroupTypes[sys], strconv.Itoa(qosLevel)); err != nil {
				return err
			}
		}
	}
	q.Log.Debugf("set pod %s(%s) qos level %d ok", pod.Name, pod.UID, qosLevel)
	return nil
}

func getQoSLevel(pod *typedef.PodInfo) int {
	if pod == nil {
		return 0
	}
	anno, ok := pod.Annotations[constant.PriorityAnnotationKey]
	if !ok {
		return 0
	}
	switch anno {
	case "true":
		return -1
	case "false":
		return 0
	default:
		return 0
	}
}

// Validate will validate the qos service config
func (q *QoS) Validate() error {
	if len(q.SubSys) == 0 {
		return fmt.Errorf("empty qos config")
	}
	for _, subSys := range q.SubSys {
		if _, ok := supportCgroupTypes[subSys]; !ok {
			return fmt.Errorf("not support sub system %s", subSys)
		}
	}
	return nil
}
