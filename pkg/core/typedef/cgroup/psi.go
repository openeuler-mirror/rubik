// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-05-18
// Description: This file is used for psi cgroup

// Package cgroup provide cgroup operations.
package cgroup

import (
	"bufio"
	"fmt"
	"strings"
)

type (
	// Stat is the cpu/mem/io.pressure data
	Stat struct {
		Avg10, Avg60, Avg300 float64
		Total                uint64
	}
	// Pressure indicates the current cgroup psi pressure
	Pressure struct {
		Some Stat
		Full Stat
	}
)

func parsePSIStats(statstype string, line string) (Stat, error) {
	var stat Stat
	if _, err := fmt.Sscanf(line, statstype+" avg10=%f avg60=%f avg300=%f total=%d",
		&stat.Avg10, &stat.Avg60, &stat.Avg300, &stat.Total); err != nil {
		return stat, fmt.Errorf("ignoring line: %s", line)
	}
	return stat, nil
}

// NewPSIData parses PSI data and returns a Pressure object
func NewPSIData(data string) (*Pressure, error) {
	/*
		cat /sys/fs/cgroup/cpuacct/io.pressure
			some avg10=0.33 avg60=3.58 avg300=2.86 total=12435574
			full avg10=0.33 avg60=2.12 avg300=1.62 total=7075471
	*/
	const (
		somePrefix = "some"
		fullPrefix = "full"
	)

	var res = Pressure{}
	scanner := bufio.NewScanner(strings.NewReader(data))
	// read by line
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, somePrefix):
			stat, err := parsePSIStats(somePrefix, line)
			if err != nil {
				return &Pressure{}, err
			}
			res.Some = stat
		case strings.HasPrefix(line, fullPrefix):
			stat, err := parsePSIStats(fullPrefix, line)
			if err != nil {
				return &Pressure{}, err
			}
			res.Full = stat
		default:
			fmt.Printf("ignoring psi line: %v\n", line)
		}
	}
	return &res, nil
}
