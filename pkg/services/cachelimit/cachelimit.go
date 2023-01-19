package cachelimit

import (
	"fmt"

	"isula.org/rubik/pkg/core/typedef"
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

type CacheLimitConfig struct {
	DefaultLimitMode  string          `json:"defaultLimitMode,omitempty"`
	DefaultResctrlDir string          `json:"-"`
	AdjustInterval    int             `json:"adjustInterval,omitempty"`
	PerfDuration      int             `json:"perfDuration,omitempty"`
	L3Percent         MultiLvlPercent `json:"l3Percent,omitempty"`
	MemBandPercent    MultiLvlPercent `json:"memBandPercent,omitempty"`
}

type CacheLimit struct {
	Name   string `json:"-"`
	Config CacheLimitConfig
}

func init() {
	services.Register("cacheLimit", func() interface{} {
		return NewCacheLimit()
	})
}

func NewCacheLimit() *CacheLimit {
	return &CacheLimit{
		Name: "cacheLimit",
		Config: CacheLimitConfig{
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
		},
	}
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

func (c *CacheLimit) ID() string {
	return c.Name
}

func (c *CacheLimit) AddFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (c *CacheLimit) UpdateFunc(old, new *typedef.PodInfo) error {
	return nil
}

func (c *CacheLimit) DeleteFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (c *CacheLimit) Validate() bool {
	return true
}
