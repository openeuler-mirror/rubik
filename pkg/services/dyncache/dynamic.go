// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2023-02-21
// Description: This file is used for dynamic cache limit level setting

// Package dyncache is the service used for cache limit setting
package dyncache

import (
	"fmt"
	"time"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// startDynamic will continuously run to detect online pod cache miss and
// limit offline pod cache usage
func (c *DynCache) startDynamic() {
	if !c.dynamicExist() {
		return
	}

	stepMore, stepLess := 5, -50
	needMore := true
	limiter := c.newCacheLimitSet(levelDynamic, c.Attr.L3PercentDynamic, c.Attr.MemBandPercentDynamic)

	for _, p := range c.listOnlinePods() {
		cacheMiss, llcMiss := getPodCacheMiss(p, c.config.PerfDuration)
		if cacheMiss >= c.Attr.MaxMiss || llcMiss >= c.Attr.MaxMiss {
			log.Infof("online pod %v cache miss: %v LLC miss: %v exceeds maxmiss, lower offline cache limit",
				p.UID, cacheMiss, llcMiss)

			if err := c.flush(limiter, stepLess); err != nil {
				log.Errorf(err.Error())
			}
			return
		}
		if cacheMiss >= c.Attr.MinMiss || llcMiss >= c.Attr.MinMiss {
			needMore = false
		}
	}

	if needMore {
		if err := c.flush(limiter, stepMore); err != nil {
			log.Errorf(err.Error())
		}
	}
}

func getPodCacheMiss(pod *typedef.PodInfo, perfDu int) (int, int) {
	cgroupPath := cgroup.AbsoluteCgroupPath("perf_event", pod.Path, "")
	if !util.PathExist(cgroupPath) {
		return 0, 0
	}

	stat, err := perf.CgroupStat(cgroupPath, time.Duration(perfDu)*time.Millisecond)
	if err != nil {
		return 0, 0
	}
	const (
		probability = 100.0
		bias        = 1.0
	)
	return int(probability * float64(stat.CacheMisses) / (bias + float64(stat.CacheReferences))),
		int(probability * float64(stat.LLCMiss) / (bias + float64(stat.LLCAccess)))
}

func (c *DynCache) dynamicExist() bool {
	for _, pod := range c.listOfflinePods() {
		if err := c.syncLevel(pod); err != nil {
			continue
		}
		if pod.Annotations[constant.CacheLimitAnnotationKey] == levelDynamic {
			return true
		}
	}
	return false
}

func (c *DynCache) flush(limitSet *limitSet, step int) error {
	var nextPercent = func(value, min, max, step int) int {
		value += step
		if value < min {
			return min
		}
		if value > max {
			return max
		}
		return value

	}
	l3 := nextPercent(c.Attr.L3PercentDynamic, c.config.L3Percent.Low, c.config.L3Percent.High, step)
	mb := nextPercent(c.Attr.MemBandPercentDynamic, c.config.MemBandPercent.Low, c.config.MemBandPercent.High, step)
	if c.Attr.L3PercentDynamic == l3 && c.Attr.MemBandPercentDynamic == mb {
		return nil
	}
	log.Infof("flush L3 from %v to %v, Mb from %v to %v", limitSet.l3Percent, l3, limitSet.mbPercent, mb)
	limitSet.l3Percent, limitSet.mbPercent = l3, mb
	return c.doFlush(limitSet)
}

func (c *DynCache) doFlush(limitSet *limitSet) error {
	if err := limitSet.writeResctrlSchemata(c.Attr.NumaNum); err != nil {
		return fmt.Errorf("adjust dynamic cache limit to l3:%v mb:%v error: %v",
			limitSet.l3Percent, limitSet.mbPercent, err)
	}
	c.Attr.L3PercentDynamic = limitSet.l3Percent
	c.Attr.MemBandPercentDynamic = limitSet.mbPercent

	return nil
}

func (c *DynCache) listOnlinePods() map[string]*typedef.PodInfo {
	onlineValue := "false"
	return c.Viewer.ListPodsWithOptions(func(pi *typedef.PodInfo) bool {
		return pi.Annotations[constant.PriorityAnnotationKey] == onlineValue
	})
}
