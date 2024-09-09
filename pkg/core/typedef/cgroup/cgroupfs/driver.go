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
// Description: This file is used for cgroupfs driver

package cgroupfs

import (
	"path/filepath"

	"isula.org/rubik/pkg/common/constant"
)

const Name = "cgroupfs"

type Driver struct{}

func (d *Driver) Name() string {
	return Name
}

func (d *Driver) ConcatPodCgroupPath(qosClass string, id string) string {
	// When using cgroupfs as cgroup driver:
	// 1. The Burstable path looks like: kubepods/burstable/pod34152897-dbaf-11ea-8cb9-0653660051c3
	// 2. The BestEffort path is in the form: kubepods/bestEffort/pod34152897-dbaf-11ea-8cb9-0653660051c3
	// 3. The Guaranteed path is in the form: kubepods/pod34152897-dbaf-11ea-8cb9-0653660051c3

	return filepath.Join(constant.KubepodsCgroup, qosClass, constant.PodCgroupNamePrefix+id)
}

func (d *Driver) GetNRIContainerCgroupPath(nriCgroupPath string) string {
	// When using cgroupfs as cgroup driver and isula, docker, containerd as container runtime:
	// 1. The Burstable path looks like: kubepods/burstable/pod34152897-dbaf-11ea-8cb9-0653660051c3/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
	// 2. The BestEffort path is in the form: kubepods/bestEffort/pod34152897-dbaf-11ea-8cb9-0653660051c3/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
	// 3. The Guaranteed path is in the form: kubepods/pod34152897-dbaf-11ea-8cb9-0653660051c3/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
	return nriCgroupPath
}

func (d *Driver) ConcatContainerCgroup(podCgroupPath, prefix, containerID string) string {
	if prefix != "" {
		prefix = prefix + "-"
	}
	return filepath.Join(podCgroupPath, prefix+containerID)
}
