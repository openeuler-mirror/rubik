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
// Date: 2023-02-08
// Description: This file is used for computing cpu utilization

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"

	"isula.org/rubik/pkg/common/util"
)

const (
	maximumUtilization float64 = 100
	minimumUtilization float64 = 0
)

// ProcStat store /proc/stat data
type ProcStat struct {
	name      string
	user      float64
	nice      float64
	system    float64
	idle      float64
	iowait    float64
	irq       float64
	softirq   float64
	steal     float64
	guest     float64
	guestNice float64
	total     float64
	busy      float64
}

// getProcStat create a proc stat object
func getProcStat() (ProcStat, error) {
	const (
		procStatFilePath   = "/proc/stat"
		nameLineNum        = 0
		userIndex          = 0
		niceIndex          = 1
		systemIndex        = 2
		idleIndex          = 3
		iowaitIndex        = 4
		irqIndex           = 5
		softirqIndex       = 6
		stealIndex         = 7
		guestIndex         = 8
		guestNiceIndex     = 9
		statsFieldsCount   = 10
		supportFieldNumber = 11
	)
	data, err := ioutil.ReadFile(procStatFilePath)
	if err != nil {
		return ProcStat{}, err
	}
	// format of the first line of the file /proc/stat :
	// name user nice system idle iowait irq softirq steal guest guest_nice
	line := strings.Fields(strings.Split(string(data), "\n")[0])
	if len(line) < supportFieldNumber {
		return ProcStat{}, fmt.Errorf("too few fields and check the kernel version")
	}
	var fields [statsFieldsCount]float64
	for i := 0; i < statsFieldsCount; i++ {
		fields[i], err = util.ParseFloat64(line[i+1])
		if err != nil {
			return ProcStat{}, err
		}
	}
	ps := ProcStat{
		name:      line[nameLineNum],
		user:      fields[userIndex],
		nice:      fields[niceIndex],
		system:    fields[systemIndex],
		idle:      fields[idleIndex],
		iowait:    fields[iowaitIndex],
		irq:       fields[irqIndex],
		softirq:   fields[softirqIndex],
		steal:     fields[stealIndex],
		guest:     fields[guestIndex],
		guestNice: fields[guestNiceIndex],
	}
	ps.busy = ps.user + ps.system + ps.nice + ps.iowait + ps.irq + ps.softirq + ps.steal
	ps.total = ps.busy + ps.idle
	return ps, nil
}

// calculateUtils calculate the CPU utilization rate based on the two interval /proc/stat
func calculateUtils(t1, t2 ProcStat) float64 {
	if t2.busy <= t1.busy {
		return minimumUtilization
	}
	if t2.total <= t1.total {
		return maximumUtilization
	}
	return math.Min(maximumUtilization,
		math.Max(minimumUtilization, util.Div(t2.busy-t1.busy, t2.total-t1.total)*maximumUtilization))
}
