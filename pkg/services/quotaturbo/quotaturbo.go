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
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	Log "isula.org/rubik/pkg/common/log"
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

// SetupLog initializes the log interface for the module
func (qt *QuotaTurbo) SetupLog(logger api.Logger) {
	log = logger
}

// NewQuotaTurbo generate quota turbo objects
func NewQuotaTurbo() *QuotaTurbo {
	return &QuotaTurbo{
		NodeData: NewNodeData(),
	}
}

// ID returns the module name
func (qt *QuotaTurbo) ID() string {
	return moduleName
}

// PreStart is the pre-start action
func (qt *QuotaTurbo) PreStart(v api.Viewer) error {
	qt.Viewer = v
	return nil
}

// saveQuota saves the quota value of the container
func (qt *QuotaTurbo) saveQuota() {
	for _, c := range qt.containers {
		if err := c.SaveQuota(); err != nil {
			log.Errorf(err.Error())
		}
	}
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
	qt.saveQuota()
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
