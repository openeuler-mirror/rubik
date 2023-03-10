package iolimit

import (
	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services/helper"
)

// DeviceConfig defines blkio device configurations
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

type IOLimitAnnoConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}

type IOLimit struct {
	helper.ServiceBase
	name string
}

type IOLimitFactory struct {
	ObjName string
}

func (i IOLimitFactory) Name() string {
	return "IOLimitFactory"
}

func (i IOLimitFactory) NewObj() (interface{}, error) {
	return &IOLimit{name: i.ObjName}, nil
}

// =========================

func (i *IOLimit) ID() string {
	return i.name
}

func (i *IOLimit) PreStart(viewer api.Viewer) error {
	log.Infof("ioLimit prestart")
	return nil
}

func (i *IOLimit) Terminate(viewer api.Viewer) error {
	log.Infof("ioLimit Terminate")
	return nil
}

func (i *IOLimit) AddFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (i *IOLimit) UpdateFunc(old, new *typedef.PodInfo) error {
	return nil
}

func (i *IOLimit) DeleteFunc(podInfo *typedef.PodInfo) error {
	return nil
}
