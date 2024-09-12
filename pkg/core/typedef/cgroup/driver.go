// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-09-03
// Description: This file is used for cgroup driver

package cgroup

import (
	"fmt"

	"isula.org/rubik/pkg/core/typedef/cgroup/cgroupfs"
	"isula.org/rubik/pkg/core/typedef/cgroup/systemd"
)

// Driver is the interface of cgroup methods
type Driver interface {
	Name() string
	ConcatPodCgroupPath(qosClass string, id string) string
	ConcatContainerCgroup(podCgroupPath, prefix, containerID string) string
	GetNRIContainerCgroupPath(nriCgroupPath string) string
}

func defaultDriver() Driver {
	return &cgroupfs.Driver{}
}

// NewCgroupDriver is the setter of global cgroup driver, only support systemd & cgroupfs
func newCgroupDriver(driverTyp string) (Driver, error) {
	switch driverTyp {
	case systemd.Name:
		return &systemd.Driver{}, nil
	case cgroupfs.Name:
		return &cgroupfs.Driver{}, nil
	}
	return nil, fmt.Errorf("invalid driver type: %v", driverTyp)
}

// Type returns the driver type
func Type() string {
	return conf.CgroupDriver.Name()
}

// ConcatPodCgroupPath returns the cgroup path of pod
func ConcatPodCgroupPath(qosClass, id string) string {
	return conf.CgroupDriver.ConcatPodCgroupPath(qosClass, id)
}

// GetNRIContainerCgroupPath returns the cgroup path of nri container
func GetNRIContainerCgroupPath(nriCgroupPath string) string {
	return conf.CgroupDriver.GetNRIContainerCgroupPath(nriCgroupPath)
}

// ConcatContainerCgroup returns the cgroup path of container from kubernetes apiserver
func ConcatContainerCgroup(podCgroupPath, prefix, containerID string) string {
	return conf.CgroupDriver.ConcatContainerCgroup(podCgroupPath, prefix, containerID)
}
