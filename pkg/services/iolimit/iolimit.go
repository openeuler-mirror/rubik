package iolimit

import (
	"isula.org/rubik/pkg/services/helper"
)

// DeviceConfig defines blkio device configurations.
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

// IOLimitAnnoConfig defines the annotation config of iolimit.
type IOLimitAnnoConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}

// IOLimit is the class of IOLimit.
type IOLimit struct {
	helper.ServiceBase
	name string
}

// IOLimitFactory is the factory of IOLimit.
type IOLimitFactory struct {
	ObjName string
}

// Name to get the IOLimit factory name.
func (i IOLimitFactory) Name() string {
	return "IOLimitFactory"
}

// NewObj to create object of IOLimit.
func (i IOLimitFactory) NewObj() (interface{}, error) {
	return &IOLimit{name: i.ObjName}, nil
}

// ID to get the name of IOLimit.
func (i *IOLimit) ID() string {
	return i.name
}
