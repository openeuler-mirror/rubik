// Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-02-11
// Description: This file provides the relevant data structures and methods of the cgroup cpu subsystem

package cgroup

import "isula.org/rubik/pkg/common/util"

type (
	// CPUStat save the cpu.stat data
	CPUStat struct {
		NrPeriods     int64
		NrThrottled   int64
		ThrottledTime int64
	}
)

// NewCPUStat creates a new MPStat object and returns its pointer
func NewCPUStat(data string) (*CPUStat, error) {
	const (
		throttlePeriodNumFieldName = "nr_periods"
		throttleNumFieldName       = "nr_throttled"
		throttleTimeFieldName      = "throttled_time"
	)
	stringInt64Map, err := util.ParseInt64Map(data)
	if err != nil {
		return nil, err
	}

	return &CPUStat{
		NrPeriods:     stringInt64Map[throttlePeriodNumFieldName],
		NrThrottled:   stringInt64Map[throttleNumFieldName],
		ThrottledTime: stringInt64Map[throttleTimeFieldName],
	}, nil
}
