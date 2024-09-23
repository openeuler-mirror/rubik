// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Kelu Ye
// Date: 2024-09-19
// Description: This file provides some methods for managing the collected data.

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"fmt"
	"sort"

	"isula.org/rubik/pkg/common/log"
)

// dataSeries represents a collected metric, consisting of timestamps and data.
type dataSeries struct {
	timeline []int64
	data     map[int64]float64
}

// newDataSeries initializes and returns an empty dataSeries with a non-nil data map.
func newDataSeries() *dataSeries {
	return &dataSeries{
		timeline: []int64{},
		data:     make(map[int64]float64),
	}
}

// add inserts a new value with a timestamp, ensuring the timestamp is non-negative and in ascending order.
func (t *dataSeries) add(value float64, nano int64) error {
	if t.data == nil {
		return fmt.Errorf("dataSeries cannot have a nil data field")
	}
	if nano < 0 {
		return fmt.Errorf("the new data timestamp cannot be negative")
	}
	if len(t.timeline) > 0 && nano < t.timeline[len(t.timeline)-1] {
		return fmt.Errorf("the new data timestamp must not be earlier than the last timestamp")
	}
	t.data[nano] = value
	t.timeline = append(t.timeline, nano)
	return nil
}

// getData search and return the value for a given timestamp, returning false if it doesn't exist.
func (t *dataSeries) getData(nano int64) (float64, bool) {
	if t.data == nil {
		return 0.0, false
	}
	v, ok := t.data[nano]
	return v, ok
}

// rangeSearch returns a new dataSeries containing data within the specified timestamp range.
func (t *dataSeries) rangeSearch(start, end int64) *dataSeries {
	if end < start {
		return nil
	}

	startIndex := sort.Search(len(t.timeline), func(i int) bool {
		return t.timeline[i] >= start
	})
	if startIndex == len(t.timeline) {
		return nil
	}

	endIndex := sort.Search(len(t.timeline), func(i int) bool {
		return t.timeline[i] >= end
	})

	if endIndex == len(t.timeline) {
		endIndex--
	}

	timeRange := t.timeline[startIndex : endIndex+1]

	ret := newDataSeries()
	for _, nano := range timeRange {
		v, ok := t.data[nano]
		if !ok {
			continue
		}
		if err := ret.add(v, nano); err != nil {
			log.Errorf("rangeSearch invoke add: %v", err)
		}
	}
	return ret
}

// normalize scales all data values so that their total sum equals 1.
func (t *dataSeries) normalize() *dataSeries {
	var sum float64
	for _, v := range t.data {
		sum += v
	}
	ret := newDataSeries()
	for _, nano := range t.timeline {
		value, ok := t.getData(nano)
		if !ok {
			continue
		}
		if err := ret.add(value/sum, nano); err != nil {
			log.Errorf("normalize invoke add: %v", err)
		}
	}
	return ret
}

// expire function removes data older than the given timestamp and returns it as a new dataSeries.
func (t *dataSeries) expire(beforeNano int64) *dataSeries {
	expiredData := newDataSeries()
	i := sort.Search(len(t.timeline), func(i int) bool {
		return t.timeline[i] >= beforeNano
	})

	expiredRange := t.timeline[:i]
	t.timeline = t.timeline[i:]
	for _, e := range expiredRange {
		if err := expiredData.add(t.data[e], e); err != nil {
			log.Errorf("expire invoke add: %v", err)
		}
		delete(t.data, e)
	}
	return expiredData
}
