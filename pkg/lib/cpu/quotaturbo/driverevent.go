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
// Description: event driver method for quota turbo

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
	"math"
	"runtime"

	"isula.org/rubik/pkg/common/util"
)

// Driver uses different methods based on different policies.
type Driver interface {
	// adjustQuota calculate the quota in the next period based on the customized policy, upper limit, and quota.
	adjustQuota(status *StatusStore)
}

// EventDriver event based quota adjustment driver.
type EventDriver struct{}

// adjustQuota calculates quota delta based on events
func (e *EventDriver) adjustQuota(status *StatusStore) {
	e.slowFallback(status)
	e.fastFallback(status)
	// Ensure that the CPU usage does not change by more than 10% within one minute.
	// Otherwise, the available quota rollback continues but does not increase.
	if !sharpFluctuates(status) {
		e.elevate(status)
	} else {
		fmt.Printf("CPU utilization fluctuates by more than %.2f\n", status.CPUFloatingLimit)
	}
	for _, c := range status.cpuQuotas {
		// get height limit
		const easingMultiple = 2.0
		// c.Period
		c.heightLimit = easingMultiple * c.cpuLimit * float64(c.period)
		// get the maximum available ensuring that the overall utilization does not exceed the limit.
		c.maxQuotaNextPeriod = getMaxQuota(c)
		// c.Period ranges from 1000(us) to 1000000(us) and does not overflow.
		c.nextQuota = int64(math.Max(math.Min(float64(c.curQuota)+c.quotaDelta, c.maxQuotaNextPeriod),
			c.cpuLimit*float64(c.period)))
	}
}

// elevate boosts when cpu is suppressed
func (e *EventDriver) elevate(status *StatusStore) {
	// the CPU usage of the current node is lower than the warning watermark.
	// U + R <= a & a > U  ======> a - U >= R && a - U > 0 =====> a - U >= R
	if float64(status.AlarmWaterMark)-status.getLastCPUUtil() < status.ElevateLimit {
		return
	}
	// sumDelta : total number of cores to be adjusted
	var sumDelta float64 = 0
	delta := make(map[string]float64, 0)
	for id, c := range status.cpuQuotas {
		if c.curThrottle.NrThrottled > c.preThrottle.NrThrottled {
			delta[id] = NsToUs(c.curThrottle.ThrottledTime-c.preThrottle.ThrottledTime) /
				float64(c.curThrottle.NrThrottled-c.preThrottle.NrThrottled) / float64(c.period)
			sumDelta += delta[id]
		}
	}
	// the container quota does not need to be increased in this round.
	if sumDelta == 0 {
		return
	}
	// the total increase cannot exceed ( status.SingleTotalIncreaseLimit% ) of the total available CPUs of the node.
	A := math.Min(sumDelta, util.PercentageToDecimal(status.ElevateLimit)*float64(runtime.NumCPU()))
	coefficient := A / sumDelta
	for id, quotaDelta := range delta {
		status.cpuQuotas[id].quotaDelta += coefficient * quotaDelta * float64(status.cpuQuotas[id].period)
	}
}

// fastFallback decreases the quota to ensure that the CPU utilization of the node is below the warning water level
// when the water level exceeds the warning water level.
func (e *EventDriver) fastFallback(status *StatusStore) {
	// The CPU usage of the current node is greater than the warning watermark, triggering a fast rollback.
	if float64(status.AlarmWaterMark) > status.getLastCPUUtil() {
		return
	}
	// sub: the total number of CPU quotas to be reduced on a node.
	sub := util.PercentageToDecimal(float64(status.AlarmWaterMark)-status.getLastCPUUtil()) * float64(runtime.NumCPU())
	// sumDelta ：total number of cpu cores that can be decreased.
	var sumDelta float64 = 0
	delta := make(map[string]float64, 0)
	for id, c := range status.cpuQuotas {
		delta[id] = float64(c.curQuota)/float64(c.period) - c.cpuLimit
		sumDelta += delta[id]
	}
	if sumDelta <= 0 {
		return
	}
	// proportional adjustment of each business quota.
	for id, quotaDelta := range delta {
		status.cpuQuotas[id].quotaDelta += (quotaDelta / sumDelta) * sub * float64(status.cpuQuotas[id].period)
	}
}

// slowFallback triggers quota callback of unpressed containers when the CPU utilization exceeds the control watermark.
func (e *EventDriver) slowFallback(status *StatusStore) {
	// The CPU usage of the current node is greater than the high watermark, triggering a slow rollback.
	if float64(status.HighWaterMark) > status.getLastCPUUtil() {
		return
	}
	coefficient := (status.getLastCPUUtil() - float64(status.HighWaterMark)) /
		float64(status.AlarmWaterMark-status.HighWaterMark) * status.SlowFallbackRatio
	for id, c := range status.cpuQuotas {
		originQuota := int64(c.cpuLimit * float64(c.period))
		if c.curQuota > originQuota && c.curThrottle.NrThrottled == c.preThrottle.NrThrottled {
			status.cpuQuotas[id].quotaDelta += coefficient * float64(originQuota-c.curQuota)
		}
	}
}

// sharpFluctuates checks whether the node CPU utilization exceeds the specified value within one minute.
func sharpFluctuates(status *StatusStore) bool {
	var (
		min float64 = maximumUtilization
		max float64 = minimumUtilization
	)
	for _, u := range status.cpuUtils {
		min = math.Min(min, u.util)
		max = math.Max(max, u.util)
	}
	return max-min > status.CPUFloatingLimit
}

// getMaxQuota calculate the maximum available quota in the next period based on the container CPU usage in N-1 periods.
func getMaxQuota(c *CPUQuota) float64 {
	if len(c.cpuUsages) <= 1 {
		return c.heightLimit
	}
	// the time unit is nanosecond
	first := c.cpuUsages[0]
	last := c.cpuUsages[len(c.cpuUsages)-1]
	timeDelta := NsToUs(last.timestamp - first.timestamp)
	coefficient := float64(len(c.cpuUsages)) / float64(len(c.cpuUsages)-1)
	maxAvailable := c.cpuLimit * timeDelta * coefficient
	used := NsToUs(last.usage - first.usage)
	remainingUsage := maxAvailable - used
	origin := c.cpuLimit * float64(c.period)
	const (
		// To prevent sharp service jitters, the Rubik proactively decreases the traffic in advance
		// when the available balance reaches a certain threshold.
		// The limitMultiplier is used to control the relationship between the upper limit and the threshold.
		// Experiments show that the value 3 is efficient and secure.
		limitMultiplier = 3
		precision       = 1e-10
	)
	var threshold = limitMultiplier * c.heightLimit
	remainingQuota := util.Div(remainingUsage, timeDelta, math.MaxFloat64, precision) *
		float64(len(c.cpuUsages)-1) * float64(c.period)

	// gradually decrease beyond the threshold to prevent sudden dips.
	res := remainingQuota
	if remainingQuota <= threshold {
		res = origin + util.Div((c.heightLimit-origin)*remainingQuota, threshold, threshold, precision)
	}
	// The utilization must not exceed the height limit and must not be less than the cpuLimit.
	return math.Max(math.Min(res, c.heightLimit), origin)
}

// NsToUs converts nanoseconds into microseconds
func NsToUs(ns int64) float64 {
	// number of nanoseconds contained in 1 microsecond
	const nanoSecPerMicroSec float64 = 1000
	return util.Div(float64(ns), nanoSecPerMicroSec)
}
