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
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/lib/cpu/quotaturbo"
	"isula.org/rubik/pkg/services/helper"
)

const (
	defaultHightWaterMark         = 60
	defaultAlarmWaterMark         = 80
	defaultQuotaTurboSyncInterval = 100
)

var (
	cpuPeriodKey = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_period_us"}
	cpuQuotaKey  = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_quota_us"}
)

// QuotaTurboFactory is the QuotaTurbo factory class
type QuotaTurboFactory struct {
	ObjName string
}

// Name returns the factory class name
func (i QuotaTurboFactory) Name() string {
	return "QuotaTurboFactory"
}

// NewObj returns a QuotaTurbo object
func (i QuotaTurboFactory) NewObj() (interface{}, error) {
	return NewQuotaTurbo(i.ObjName), nil
}

// Config is the config of QuotaTurbo
type Config struct {
	HighWaterMark  int `json:"highWaterMark,omitempty"`
	AlarmWaterMark int `json:"alarmWaterMark,omitempty"`
	SyncInterval   int `json:"syncInterval,omitempty"`
}

// NewConfig returns quotaTurbo config instance
func NewConfig() *Config {
	return &Config{
		HighWaterMark:  defaultHightWaterMark,
		AlarmWaterMark: defaultAlarmWaterMark,
		SyncInterval:   defaultQuotaTurboSyncInterval,
	}
}

// QuotaTurbo manages all container CPU quota data on the current node.
type QuotaTurbo struct {
	conf   *Config
	client quotaturbo.ClientAPI
	Viewer api.Viewer
	helper.ServiceBase
}

// NewQuotaTurbo generate quota turbo objects
func NewQuotaTurbo(n string) *QuotaTurbo {
	return &QuotaTurbo{
		ServiceBase: helper.ServiceBase{
			Name: n,
		},
		conf:   NewConfig(),
		client: quotaturbo.NewClient(),
	}
}

// syncCgroups updates the cgroup in cilent according to the current whitelist pod list
func (qt *QuotaTurbo) syncCgroups(conts map[string]*typedef.ContainerInfo) {
	var (
		existedCgroupPaths   = qt.client.AllCgroups()
		existedCgroupPathMap = make(map[string]struct{}, len(existedCgroupPaths))
	)
	// delete containers marked as no need to adjust quota
	for _, path := range existedCgroupPaths {
		id := filepath.Base(path)
		existedCgroupPathMap[id] = struct{}{}
		if _, found := conts[id]; !found {
			if err := qt.client.RemoveCgroup(path); err != nil {
				log.Errorf("failed to remove container %v: %v", id, err)
			} else {
				log.Infof("remove container %v successfully", id)
			}
		}
	}
	for id, cont := range conts {
		// Currently, modifying the cpu limit and container id of the container will cause the container to restart,
		// so it is considered that the cgroup path and cpulimit will not change during the life cycle of the container
		if _, ok := existedCgroupPathMap[id]; ok {
			continue
		}
		// add container to quotaturbo
		if err := qt.client.AddCgroup(cont.Path, cont.LimitResources[typedef.ResourceCPU]); err != nil {
			log.Errorf("failed to add container %v: %v", cont.Name, err)
		} else {
			log.Infof("add container %v successfully", id)
		}
	}
}

// AdjustQuota adjusts the quota of a container at a time
func (qt *QuotaTurbo) AdjustQuota(conts map[string]*typedef.ContainerInfo) {
	qt.syncCgroups(conts)
	if err := qt.client.AdjustQuota(); err != nil {
		log.Errorf("an error occurred while adjusting the quota: %v", err)
	}
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
		time.Millisecond*time.Duration(qt.conf.SyncInterval),
		ctx.Done())
}

// Validate verifies that the quotaTurbo parameter is set correctly
func (conf *Config) Validate() error {
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
	if conf.AlarmWaterMark <= conf.HighWaterMark ||
		outOfRange(conf.HighWaterMark, minQuotaTurboWaterMark, maxQuotaTurboWaterMark) ||
		outOfRange(conf.AlarmWaterMark, minQuotaTurboWaterMark, maxQuotaTurboWaterMark) {
		return fmt.Errorf("alarmWaterMark >= highWaterMark, both of which ranges from 0 to 100")
	}
	if outOfRange(conf.SyncInterval, minQuotaTurboSyncInterval, maxQuotaTurboSyncInterval) {
		return fmt.Errorf("synchronization time ranges from 100 (0.1s) to 10000 (10s)")
	}
	return nil
}

// SetConfig sets and checks Config
func (qt *QuotaTurbo) SetConfig(f helper.ConfigHandler) error {
	if f == nil {
		return fmt.Errorf("no config handler function callback")
	}

	var conf = NewConfig()
	if err := f(qt.Name, conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}
	qt.conf = conf
	return nil
}

// GetConfig returns Config
func (qt *QuotaTurbo) GetConfig() interface{} {
	return qt.conf
}

// IsRunner returns true that tells other quotaTurbo is a persistent service
func (qt *QuotaTurbo) IsRunner() bool {
	return true
}

// PreStart is the pre-start action
func (qt *QuotaTurbo) PreStart(viewer api.Viewer) error {
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	// 1. set the parameters of the quotaturbo client
	qt.client.WithOptions(
		quotaturbo.WithCgroupRoot(cgroup.GetMountDir()),
		quotaturbo.WithWaterMark(qt.conf.HighWaterMark, qt.conf.AlarmWaterMark),
	)
	qt.Viewer = viewer

	// 2. attempts to fix all currently running pods and containers
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
		log.Errorf("failed to set the cpu quota of the pod %v to unlimited(-1): %v", pod.UID, err)
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
				log.Errorf("failed to set the cpu quota of the container %v to unlimited(-1): %v", cont.ID, err)
				continue
			}
			log.Debugf("set the cpu quota of the container %v to unlimited(-1)", cont.ID)
			continue
		}

		period, err := cont.GetCgroupAttr(cpuPeriodKey).Int64()
		if err != nil {
			log.Errorf("failed to get cpu period of container %v : %v", cont.ID, err)
			continue
		}
		// the value range of cpu.cfs_period_us is 1000 (1ms) to 1000000 (1s),
		// the number of CPUs configured to the container will not exceed the number of physical machine cores
		contQuota := int64(cont.LimitResources[typedef.ResourceCPU] * float64(period))
		podQuota += contQuota
		if err := cont.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(contQuota)); err != nil {
			log.Errorf("failed to set the cpu quota of the container %v: %v", cont.ID, err)
			continue
		}
		log.Debugf("set the cpu quota of the container %v to %v", cont.ID, contQuota)
	}
	if !unlimitedContExistd {
		if err := pod.SetCgroupAttr(cpuQuotaKey, util.FormatInt64(podQuota)); err != nil {
			log.Errorf("failed to set the cpu quota of the pod %v to unlimited(-1): %v", pod.UID, err)
			return
		}
		log.Debugf("set the cpu quota of the pod %v to %v", pod.UID, podQuota)
	}
}
