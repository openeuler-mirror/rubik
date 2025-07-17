// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Niu Qianqian
// Create: 2025-07-01
// Description: Integrating oncn-bwm features

package preemption

import (
	"fmt"
	"strconv"
	"strings"

	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	// Enable netqos in the specified process network namespace
	// Usage:
	// 		echo $pid > /proc/qos/net_qos_enable
	netQosEnablePath = "/proc/qos/net_qos_enable"
	// Disable netqos in the specified process network namespace
	// Usage:
	// 		echo $pid > /proc/qos/net_qos_disable
	netQosDisablePath = "/proc/qos/net_qos_disable"
	// Set/get the bandwidth of offline pod
	// Usage:
	// 		echo "$low,$high" > /proc/qos/net_qos_bandwidth
	//		cat /proc/qos/net_qos_bandwidth
	netQosBandwidthPath = "/proc/qos/net_qos_bandwidth"
	// Set/get the waterline of offline pod
	// Usage:
	// 		echo "$val" > /proc/qos/net_qos_waterline
	//		cat /proc/qos/net_qos_waterline
	netQosWaterlinePath = "/proc/qos/net_qos_waterline"
)

func isSupportNetqos() bool {
	return util.PathExist(netQosEnablePath)
}

func getPodProcID(pod *typedef.PodInfo) (string, error) {
	var procID string
	cgroupKey := &cgroup.Key{SubSys: "net_cls", FileName: "cgroup.procs"}
	for _, container := range pod.IDContainersMap {
		key := container.GetCgroupAttr(cgroupKey)
		if key.Err != nil {
			continue
		}
		procID = strings.Split(key.Value, "\n")[0]
		procIDInt, err := strconv.Atoi(procID)
		if err == nil && procIDInt != 0 {
			return procID, nil
		}
	}

	return "", fmt.Errorf("failed to find valid proc")
}

func enablePodNetqos(pid string) error {
	if err := util.WriteFile(netQosEnablePath, pid); err != nil {
		return fmt.Errorf("failed to write %s to file %s: %v", pid, netQosEnablePath, err)
	}
	return nil
}

func disablePodNetqos(pid string) error {
	if err := util.WriteFile(netQosDisablePath, pid); err != nil {
		return fmt.Errorf("failed to write %s to file %s: %v", pid, netQosDisablePath, err)
	}
	return nil
}

func setPodNetqosBandwidth(bandwidthLow, bandwidthHigh int) error {
	bandwidthStr := strconv.Itoa(bandwidthLow) + "mb," + strconv.Itoa(bandwidthHigh) + "mb"

	if err := util.WriteFile(netQosBandwidthPath, bandwidthStr); err != nil {
		return fmt.Errorf("failed to write %s to file %s: %v", bandwidthStr, netQosBandwidthPath, err)
	}
	return nil
}

func setPodNetqosWaterline(waterline int) error {
	waterlineStr := strconv.Itoa(waterline) + "mb"

	if err := util.WriteFile(netQosWaterlinePath, waterlineStr); err != nil {
		return fmt.Errorf("failed to write %s to file %s: %v", waterlineStr, netQosWaterlinePath, err)
	}
	return nil
}
