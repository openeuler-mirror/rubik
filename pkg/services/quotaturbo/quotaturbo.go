// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-02-16
// Description: quota turbo method（dynamically adjusting container quotas）

// Package quotaturbo is for Quota Turbo
package quotaturbo

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	Log "isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/services"
)

const moduleName = "quotaturbo"

var log api.Logger

func init() {
	log = &Log.EmptyLog{}
	services.Register(moduleName, func() interface{} {
		return NewQuotaTurbo()
	})
}

// QuotaTurbo manages all container CPU quota data on the current node.
type QuotaTurbo struct {
	// NodeData including the container data, CPU usage, and QuotaTurbo configuration of the local node
	*NodeData
	// interfaces with different policies
	Driver
	// referenced object to list pods
	Viewer api.Viewer
}

// NewQuotaTurbo generate quota turbo objects
func NewQuotaTurbo() *QuotaTurbo {
	return &QuotaTurbo{
		NodeData: NewNodeData(),
	}
}

// SetupLog initializes the log interface for the module
func (qt *QuotaTurbo) SetupLog(logger api.Logger) {
	log = logger
}

// ID returns the module name
func (qt *QuotaTurbo) ID() string {
	return moduleName
}

// AdjustQuota adjusts the quota of a container at a time
func (qt *QuotaTurbo) AdjustQuota(cc map[string]*typedef.ContainerInfo) {
	qt.UpdateClusterContainers(cc)
	if err := qt.updateCPUUtils(); err != nil {
		log.Errorf("fail to get current cpu utilization : %v", err)
		return
	}
	if len(qt.containers) == 0 {
		return
	}
	qt.adjustQuota(qt.NodeData)
	qt.WriteQuota()
}

// Run adjusts the quota of the trust list container cyclically.
func (qt *QuotaTurbo) Run(ctx context.Context) {
	wait.Until(
		func() {
			qt.AdjustQuota(qt.Viewer.ListContainersWithOptions(
				func(pod *typedef.PodInfo) bool {
					return pod.Annotations[constant.QuotaAnnotationKey] == "true"
				}))
		},
		time.Millisecond*time.Duration(qt.SyncInterval),
		ctx.Done())
}

// Validate Validate verifies that the quotaTurbo parameter is set correctly
func (qt *QuotaTurbo) Validate() error {
	const (
		minQuotaTurboWaterMark, maxQuotaTurboWaterMark       = 0, 100
		minQuotaTurboSyncInterval, maxQuotaTurboSyncInterval = 100, 10000
	)
	outOfRange := func(num, min, max int) bool {
		if num < min || num > max {
			return true
		}
		return false
	}
	if qt.AlarmWaterMark <= qt.HighWaterMark ||
		outOfRange(qt.HighWaterMark, minQuotaTurboWaterMark, maxQuotaTurboWaterMark) ||
		outOfRange(qt.AlarmWaterMark, minQuotaTurboWaterMark, maxQuotaTurboWaterMark) {
		return fmt.Errorf("alarmWaterMark >= highWaterMark, both of which ranges from 0 to 100")
	}
	if outOfRange(qt.SyncInterval, minQuotaTurboSyncInterval, maxQuotaTurboSyncInterval) {
		return fmt.Errorf("synchronization time ranges from 100 (0.1s) to 10000 (10s)")
	}
	return nil
}

// PreStart is the pre-start action
func (qt *QuotaTurbo) PreStart(viewer api.Viewer) error {
	qt.Viewer = viewer
	pods := viewer.ListPodsWithOptions()
	for _, pod := range pods {
		recoverOnePodQuota(pod)
	}
	return nil
}

// Terminate enters the service termination process
func (qt *QuotaTurbo) Terminate(viewer api.Viewer) error {
	pods := viewer.ListPodsWithOptions()
	for _, pod := range pods {
		recoverOnePodQuota(pod)
	}
	return nil
}

func recoverOnePodQuota(pod *typedef.PodInfo) {
	const unlimited = "-1"
	if err := pod.SetCgroupAttr(cpuQuotaKey, unlimited); err != nil {
		log.Errorf("Fail to set the cpu quota of the pod %v to -1: %v", pod.UID, err)
		return
	}

	var (
		podQuota            int64 = 0
		unlimitedContExistd       = false
	)

	for _, cont := range pod.IDContainersMap {
		// cpulimit is 0 means no quota limit
		if cont.LimitResources[typedef.ResourceCPU] == 0 {
			unlimitedContExistd = true
			if err := cont.SetCgroupAttr(cpuQuotaKey, unlimited); err != nil {
				log.Errorf("Fail to set the cpu quota of the container %v to -1: %v", cont.ID, err)
				continue
			}
			log.Debugf("Set the cpu quota of the container %v to -1", cont.ID)
			continue
		}

		period, err := cont.GetCgroupAttr(cpuPeriodKey).Int64()
		if err != nil {
			log.Errorf("Fail to get cpu period of container %v : %v", cont.ID, err)
			continue
		}

		contQuota := int64(cont.LimitResources[typedef.ResourceCPU] * float64(period))
		podQuota += contQuota
		if err := cont.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(contQuota)); err != nil {
			log.Errorf("Fail to set the cpu quota of the container %v: %v", cont.ID, err)
			continue
		}
		log.Debugf("Set the cpu quota of the container %v to %v", cont.ID, contQuota)
	}
	if !unlimitedContExistd {
		if err := pod.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(podQuota)); err != nil {
			log.Errorf("Fail to set the cpu quota of the pod %v to -1: %v", pod.UID, err)
			return
		}
		log.Debugf("Set the cpu quota of the pod %v to %v", pod.UID, podQuota)
	}
}
