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
// Description: This file is used for testing dataSeries.go

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"math"
	"testing"
	"time"
)

const epsilon = 1e-9

// TestNewDataSeries tests NewDataSeries function
func TestNewDataSeries(t *testing.T) {
	t.Run("test newDataSeries", func(t *testing.T) {
		if got := newDataSeries(); got == nil {
			t.Errorf("newDataSeries() returns nil")
		} else if got.timeline == nil || got.data == nil {
			t.Errorf("newDataSeries() hasn't been initialized correctly")
		}
	})
}

// TestDataSeriesAdd tests Add function
func TestDataSeriesAdd(t *testing.T) {
	testSeries := newDataSeries()
	type args struct {
		value float64
		nano  int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nano is negative",
			args: args{
				value: 0.1,
				nano:  -1 * time.Now().UnixNano(),
			},
			wantErr: true,
		},
		{
			name: "valid timestamp",
			args: args{
				value: 0.1,
				nano:  time.Now().UnixNano(),
			},
			wantErr: false,
		},
		{
			name: "new timestamp is earlier than the previous timestamp",
			args: args{
				value: 0.1,
				nano:  time.Now().Add(-1 * time.Minute).UnixNano(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testSeries.add(tt.args.value, tt.args.nano); (err != nil) != tt.wantErr {
				t.Fatalf("testSeries.add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				lastKey := testSeries.timeline[len(testSeries.timeline)-1]
				lastValue, _ := testSeries.getData(lastKey)
				if lastKey != tt.args.nano {
					t.Errorf("testSeries.add() added an incorrect key: got %v, want %v", lastKey, tt.args.nano)
				}
				if lastValue != tt.args.value {
					t.Errorf("testSeries.add() added an incorrect value: got %v, want %v", lastValue, tt.args.value)
				}
			}
		})
	}
}

// TestDataSeriesRangeSearch tests RangeSearch function
func TestDataSeriesRangeSearch(t *testing.T) {
	testSeries, now := generateTestSeries()
	type args struct {
		start int64
		end   int64
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name: "startTime is greater than endTime",
			args: args{
				start: now.Add(-4 * time.Minute).UnixNano(),
				end:   now.Add(-6 * time.Minute).UnixNano(),
			},
			wantNil: true,
		},
		{
			name: "start is greater than all timeStamp",
			args: args{
				start: now.Add(time.Minute).UnixNano(),
				end:   now.UnixNano(),
			},
			wantNil: true,
		},
		{
			name: "end is smaller than all timeStamp",
			args: args{
				start: now.Add(time.Minute).UnixNano(),
				end:   now.Add(-10 * time.Minute).UnixNano(),
			},
			wantNil: true,
		},
		{
			name: "everything is fine",
			args: args{
				start: now.Add(-6 * time.Minute).UnixNano(),
				end:   now.Add(-4 * time.Minute).UnixNano(),
			},
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testSeries.rangeSearch(tt.args.start, tt.args.end)
			if (got == nil) != tt.wantNil {
				t.Errorf("testSeries.rangeSearch() = %v, want %v", got, tt.wantNil)
			}
			if got != nil && tt.wantNil == false {
				targetSeries := newDataSeries()
				for i := 9; i >= 0; i-- {
					timePoint := now.Add(time.Duration(-i) * time.Minute).UnixNano()
					if (timePoint >= tt.args.start) && (timePoint <= tt.args.end) {
						targetSeries.add(float64(i), timePoint)
					}
				}
				if !dataSeriesEqual(got, targetSeries) {
					t.Fatalf("the value of expire errors")
				}
			}
		})
	}
}

// TestDataSeriesNormalize tests Normalize function
func TestDataSeriesNormalize(t *testing.T) {
	testSeries, now := generateTestSeries()
	tests := []struct {
		name string
		want *dataSeries
	}{
		{
			name: "everything is fine",
			want: newDataSeries(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testSeries.normalize()
			var sum float64
			for i := 9; i >= 0; i-- {
				sum += float64(i)
			}
			now = now.Add(-10 * time.Minute)
			for i := 9; i >= 0; i-- {
				now = now.Add(time.Minute)
				tt.want.add(float64(i)/sum, now.UnixNano())
			}
			if !dataSeriesEqual(got, tt.want) {
				t.Fatalf("the value of expire errors")
			}
		})
	}
}

// TestDataSeriesExpire tests Expire function
func TestDataSeriesExpire(t *testing.T) {
	testSeries, now := generateTestSeries()
	tempStore := cloneDataSeries(testSeries)
	tests := []struct {
		name       string
		beforeNano int64
		want       *dataSeries
	}{
		{
			name:       "all expired",
			beforeNano: now.Add(time.Minute).UnixNano(),
			want:       newDataSeries(),
		},
		{
			name:       "nothing expired",
			beforeNano: now.Add(-10 * time.Minute).UnixNano(),
			want:       newDataSeries(),
		},
		{
			name:       "partially expired",
			beforeNano: now.Add(-4 * time.Minute).UnixNano(),
			want:       newDataSeries(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, nano := range testSeries.timeline {
				if nano < tt.beforeNano {
					value, ok := testSeries.getData(nano)
					if !ok {
						continue
					}
					tt.want.add(value, nano)
				} else {
					break
				}
			}
			got := testSeries.expire(tt.beforeNano)

			if !dataSeriesEqual(got, tt.want) {
				t.Fatalf("expire function returned incorrect values")
			}
			testSeries = cloneDataSeries(tempStore)
		})
	}
}

// generateTestSeries is used to generate dataSeries for test
func generateTestSeries() (*dataSeries, time.Time) {
	now := time.Now().Add(-10 * time.Minute)
	testSeries := newDataSeries()
	for i := 9; i >= 0; i-- {
		now = now.Add(time.Minute)
		testSeries.add(float64(i), now.UnixNano())
	}
	return testSeries, now
}

// cloneDataSeries creates a deep copy of the given dataSeries, replicating its timeline and data values.
func cloneDataSeries(origin *dataSeries) *dataSeries {
	res := newDataSeries()
	for _, nano := range origin.timeline {
		value, ok := origin.getData(nano)
		if !ok {
			continue
		}
		res.add(value, nano)
	}
	return res
}

// dataSeriesEqual compares two dataSeries for equality, checking if both have the same timeline and data values within a small tolerance.
func dataSeriesEqual(first *dataSeries, second *dataSeries) bool {
	if first == nil || second == nil {
		if first == nil && second == nil {
			return true
		} else {
			return false
		}
	}

	if len(first.timeline) != len(second.timeline) {
		return false
	}
	for _, nano := range second.timeline {
		secondValue, _ := second.getData(nano)
		firstValue, ok := first.getData(nano)
		if !ok || math.Abs(firstValue-secondValue) > epsilon {
			return false
		}
	}
	return true
}
