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
	"isula.org/rubik/pkg/core/typedef/cgroup/cgroupfs"
	"isula.org/rubik/pkg/core/typedef/cgroup/systemd"
)

type Driver interface {
	Name() string
	ConcatPodCgroupPath(qosClass string, id string) string
	ConcatContainerCgroup(podCgroupPath, prefix, containerID string) string
	GetNRIContainerCgroupPath(nriCgroupPath string) string
}

var driver Driver = &cgroupfs.Driver{}

// SetCgroupDriver is the setter of global cgroup driver
func SetCgroupDriver(driverTyp string) {
	cgroupDriver = driverTyp
	switch driverTyp {
	case systemd.Name:
		driver = &systemd.Driver{}
	case cgroupfs.Name:
		driver = &cgroupfs.Driver{}
	}
}

func Type() string {
	return driver.Name()
}
func ConcatPodCgroupPath(qosClass, id string) string {
	return driver.ConcatPodCgroupPath(qosClass, id)
}

func GetNRIContainerCgroupPath(nriCgroupPath string) string {
	return driver.GetNRIContainerCgroupPath(nriCgroupPath)
}

func ConcatContainerCgroup(podCgroupPath, prefix, containerID string) string {
	return driver.ConcatContainerCgroup(podCgroupPath, prefix, containerID)
}
