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

const (
	Name          = "systemd"
	cgroupFileExt = ".scope"
)

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

func (d *Driver) GetNRIContainerCgroupPath(nriCgroupPath string) string {
	// When using systemd as cgroup driver:
	// 1. The Burstable path looks like:
	// kubepods.slice/kubepods-burstable.slice/kubepods-burstable-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec.scope
	// 2. The BestEffort path is in the form:
	// kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec.scope
	// 3. The Guaranteed path is in the form:
	// kubepods.slice/kubepods-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec.scope

	// what we get from nri container: kubepods-besteffort-pod7631cab3_4785_4a70_a4f3_03505fb28b64.slice:cri-containerd:d4d54e90e1c55e71910e5196d4e65be39ff8c5fb39a2c3e662893ff2ab9b42cd

	// 1. we parse cgroupPath and get
	// parent: kubepods-besteffort-pod7631cab3_4785_4a70_a4f3_03505fb28b64.slice
	// prefix: cri-containerd
	// containerID: d4d54e90e1c55e71910e5196d4e65be39ff8c5fb39a2c3e662893ff2ab9b42cd
	parts := strings.Split(nriCgroupPath, ":")
	var parent, prefix, id = parts[0], parts[1], parts[2]

	// 2. the last segment of the path must be cri-containerd-d4d54e90e1c55e71910e5196d4e65be39ff8c5fb39a2c3e662893ff2ab9b42cd.scope
	last := prefix + "-" + id + cgroupFileExt

	// 3. parse parent to obtain the upper directory
	parentParts := strings.Split(parent, "-")
	var upperDir string
	// for the Guaranteed, we get 2 parts: kubepods, pod9d8d5026_5f11_4530_b929_c11f833027c2.slice
	// for the others, we get 3 parts: kubepods, besteffort, pod7631cab3_4785_4a70_a4f3_03505fb28b64.slice
	if len(parentParts) > 2 {
		// This means upper directory should contain a kubepods-besteffort.slice or kubepods-burstable.slice
		upperDir = parentParts[0] + "-" + parentParts[1] + ".slice"
	}

	const topDir = constant.KubepodsCgroup + ".slice"
	return filepath.Join(topDir, upperDir, parent, last)
}

func (d *Driver) ConcatContainerCgroup(podCgroupPath, prefix, containerID string) string {
	if prefix != "" {
		prefix = prefix + "-"
	}
	return filepath.Join(podCgroupPath, prefix+containerID+cgroupFileExt)
}
