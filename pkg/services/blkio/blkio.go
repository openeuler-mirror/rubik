package blkio

import (
	"fmt"

	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services"
)

// DeviceConfig defines blkio device configurations
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

type BlkioConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}

type Blkio struct {
	Name   string `json:"-"`
	Config BlkioConfig
}

func init() {
	services.Register("blkio", func() interface{} {
		return NewBlkio()
	})
}

func NewBlkio() *Blkio {
	return &Blkio{Name: "blkio"}
}

func (b *Blkio) Setup() error {
	fmt.Println("blkio Setup()")
	return nil
}

func (b *Blkio) TearDown() error {
	fmt.Println("blkio TearDown()")
	return nil
}

func (b *Blkio) ID() string {
	return b.Name
}

func (b *Blkio) AddFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (b *Blkio) UpdateFunc(old, new *typedef.PodInfo) error {
	return nil
}

func (b *Blkio) DeleteFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (b *Blkio) Validate() bool {
	return true
}
