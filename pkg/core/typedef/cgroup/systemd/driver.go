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
// Description: This file is used for system cgroup driver

package systemd

import (
	"path/filepath"
	"strings"

	"isula.org/rubik/pkg/common/constant"
)

const Name = "systemd"

type Driver struct{}

func (d *Driver) Name() string {
	return Name
}

func (d *Driver) ConcatPodCgroupPath(qosClass string, id string) string {
	// When using systemd as cgroup driver:
	// 1. The Burstable path looks like: kubepods.slice/kubepods-burstable.slice/kubepods-burstable-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice
	// 2. The BestEffort path is in the form: kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice
	// 3. The Guaranteed path is in the form: kubepods.slice/kubepods-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/
	const suffix = ".slice"
	var (
		prefix  = constant.KubepodsCgroup
		podPath = constant.KubepodsCgroup + suffix
	)
	if qosClass != "" {
		podPath = filepath.Join(podPath, constant.KubepodsCgroup+"-"+qosClass+suffix)
		prefix = strings.Join([]string{prefix, qosClass}, "-")
	}
	return filepath.Join(podPath,
		strings.Join([]string{prefix, constant.PodCgroupNamePrefix + strings.Replace(id, "-", "_", -1) + suffix}, "-"))
}

func (d *Driver) ConcatContainerCgroupPath(podCgroupPath string, containerScope string) string {
	return filepath.Join(podCgroupPath, containerScope+".scope")
}
