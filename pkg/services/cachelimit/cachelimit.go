package cachelimit

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"isula.org/rubik/pkg/services"
)

const (
	defaultLogSize = 1024
	defaultAdInt   = 1000
	defaultPerfDur = 1000
	defaultLowL3   = 20
	defaultMidL3   = 30
	defaultHighL3  = 50
	defaultLowMB   = 10
	defaultMidMB   = 30
	defaultHighMB  = 50
)

// MultiLvlPercent define multi level percentage
type MultiLvlPercent struct {
	Low  int `json:"low,omitempty"`
	Mid  int `json:"mid,omitempty"`
	High int `json:"high,omitempty"`
}

type CacheLimit struct {
	Name              string          `json:"-"`
	DefaultLimitMode  string          `json:"defaultLimitMode,omitempty"`
	DefaultResctrlDir string          `json:"-"`
	AdjustInterval    int             `json:"adjustInterval,omitempty"`
	PerfDuration      int             `json:"perfDuration,omitempty"`
	L3Percent         MultiLvlPercent `json:"l3Percent,omitempty"`
	MemBandPercent    MultiLvlPercent `json:"memBandPercent,omitempty"`
}

func init() {
	services.Register("cacheLimit", func() interface{} {
		return NewCacheLimit()
	})
}

func NewCacheLimit() *CacheLimit {
	return &CacheLimit{
		DefaultLimitMode:  "static",
		DefaultResctrlDir: "/sys/fs/resctrl",
		AdjustInterval:    defaultAdInt,
		PerfDuration:      defaultPerfDur,
		L3Percent: MultiLvlPercent{
			Low:  defaultLowL3,
			Mid:  defaultMidL3,
			High: defaultHighL3,
		},
		MemBandPercent: MultiLvlPercent{
			Low:  defaultLowMB,
			Mid:  defaultMidMB,
			High: defaultHighMB,
		},
	}
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
	return nil
}
