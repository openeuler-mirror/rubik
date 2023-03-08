package blkcg

import (
	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services"
)

// DeviceConfig defines blkio device configurations
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

type BlkioThrottleConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}

type BlkioThrottle struct {
	Name string `json:"-"`
	Log  api.Logger
}

func init() {
	services.Register("blkio", func() interface{} {
		return NewBlkioThrottle()
	})
}

func NewBlkioThrottle() *BlkioThrottle {
	return &BlkioThrottle{Name: "blkiothrottle"}
}

func (b *BlkioThrottle) PreStart(viewer api.Viewer) error {
	b.Log.Debugf("blkiothrottle prestart")
	return nil
}

func (b *BlkioThrottle) Terminate(viewer api.Viewer) error {
	b.Log.Infof("blkiothrottle Terminate")
	return nil
}

func (b *BlkioThrottle) ID() string {
	return b.Name
}

func (b *BlkioThrottle) AddFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (b *BlkioThrottle) UpdateFunc(old, new *typedef.PodInfo) error {
	return nil
}

func (b *BlkioThrottle) DeleteFunc(podInfo *typedef.PodInfo) error {
	return nil
}
