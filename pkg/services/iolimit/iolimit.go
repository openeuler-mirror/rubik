// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2023-03-11
// Description: This file is used to implement iolimit

// Package iolimit
package iolimit

import (
	"isula.org/rubik/pkg/services/helper"
)

// DeviceConfig defines blkio device configurations.
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

// annoConfig defines the annotation config of iolimit.
/*
type annoConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}
*/

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
