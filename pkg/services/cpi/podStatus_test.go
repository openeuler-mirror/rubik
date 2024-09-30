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
// Description: This file is used for testing podstatus.go

// Package cpi is for CPU Interference Detection Service
package cpi

import (
	"math"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// TestPodStatusUpdateThreshold tests updateThreshold function
func TestPodStatusUpdateThreshold(t *testing.T) {
	type fields struct {
		isOnline bool
		cpiMean  float64
		count    int64
		stdDev   float64
	}
	type target struct {
		targetMean   float64
		targetStdDev float64
		targetCount  int64
	}
	tests := []struct {
		name    string
		fields  fields
		expired *dataSeries
		target  target
	}{
		{
			name: "pre count = 0",
			fields: fields{
				isOnline: true,
				cpiMean:  0.0,
				count:    0,
				stdDev:   0.0,
			},
			expired: &dataSeries{
				timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				data: map[int64]float64{
					1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
				},
			},
			target: target{
				targetMean:   2.118181818181818,
				targetCount:  11,
				targetStdDev: 0.1898237547074646,
			},
		},
		{
			name: "pre count = 5",
			fields: fields{
				isOnline: true,
				cpiMean:  2.0,
				count:    5,
				stdDev:   0.1,
			},
			expired: &dataSeries{
				timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				data: map[int64]float64{
					1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
				},
			},
			target: target{
				targetMean:   2.08125,
				targetCount:  16,
				targetStdDev: 0.17577951388031546,
			},
		},
		{
			name: "pod is offlinePod",
			fields: fields{
				isOnline: false,
				cpiMean:  0.0,
				count:    0,
				stdDev:   0.0,
			},
			expired: &dataSeries{
				timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				data: map[int64]float64{
					1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
				},
			},
			target: target{
				targetMean:   0.0,
				targetCount:  0,
				targetStdDev: 0.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podStatus := &podStatus{
				isOnline: tt.fields.isOnline,
				cpiMean:  tt.fields.cpiMean,
				count:    tt.fields.count,
				stdDev:   tt.fields.stdDev,
			}
			if !podStatus.isOnline {
				return
			}
			podStatus.updateCPIStatistic(tt.expired)
			if math.Abs(podStatus.cpiMean-tt.target.targetMean) > epsilon || math.Abs(podStatus.stdDev-tt.target.targetStdDev) > epsilon || podStatus.count != tt.target.targetCount {
				t.Error("updateThreshold result errors")
			}
		})
	}
}

// TestPodStatusExpireAndUpdateThreshold tests expireAndUpdateThreshold function
func TestPodStatusExpireAndUpdateThreshold(t *testing.T) {
	_, err := perf.CgroupStat(cgroup.AbsoluteCgroupPath("perf_event", "", ""), time.Millisecond, cpiConf)
	if err != nil {
		return
	}
	type fields struct {
		isOnline       bool
		cpiSeries      *dataSeries
		cpuUsageSeries *dataSeries
		cpiMean        float64
		count          int64
		stdDev         float64
	}
	type target struct {
		targetMean      float64
		targetStdDev    float64
		targetCpiSeries *dataSeries
		targetCount     int64
	}
	tests := []struct {
		name       string
		fields     fields
		expireNano int64
		target     target
	}{{
		name: "offlinePod only expire cpuUsage",
		fields: fields{
			isOnline: false,
			cpuUsageSeries: &dataSeries{
				timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				data: map[int64]float64{
					1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
				},
			},
			cpiMean: 0,
			count:   0,
			stdDev:  0,
		},
		expireNano: 4,
	},
		{
			name: "onlinePod only expire",
			fields: fields{
				isOnline: true,
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
					},
				},
				cpiMean: 0,
				count:   0,
				stdDev:  0,
			},
			expireNano: 4,
			target: target{
				targetMean:   2.0666666666666664,
				targetStdDev: 0.047140452079103216,
				targetCount:  3,
			},
		},
		{
			name: "onlinePod expires and flush cpuUsage < 0.25",
			fields: fields{
				isOnline: true,
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 0.20, 2: 2.1, 3: 2.0, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.0, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
					},
				},
				cpiMean: 0,
				count:   0,
				stdDev:  0,
			},
			expireNano: 4,
			target: target{
				targetMean:   2.05,
				targetStdDev: 0.05,
				targetCount:  2,
			},
		},
		{
			name: "onlinePod expires and flush cpi > cpiMean + 2 * stdDev",
			fields: fields{
				isOnline: true,
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.0, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 4, 2: 2.0, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.1, 11: 2.7,
					},
				},
				cpiMean: 2,
				count:   10,
				stdDev:  0.5,
			},
			expireNano: 4,
			target: target{
				targetMean:   2.0083333333333333,
				targetStdDev: 0.4572714972772983,
				targetCount:  12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podStatus := &podStatus{
				isOnline:       tt.fields.isOnline,
				cpiSeries:      tt.fields.cpiSeries,
				cpuUsageSeries: tt.fields.cpuUsageSeries,
				cpiMean:        tt.fields.cpiMean,
				count:          tt.fields.count,
				stdDev:         tt.fields.stdDev,
			}
			targetPodStatus := podStatus.clone()
			targetPodStatus.cpuUsageSeries.expire(tt.expireNano)

			podStatus.expireAndUpdateThreshold(tt.expireNano)

			if !dataSeriesEqual(targetPodStatus.cpuUsageSeries, podStatus.cpuUsageSeries) {
				t.Errorf("expireAndUpdateThreshold cpuUsageSeries errors")
			}
			if tt.target.targetCount != podStatus.count || math.Abs(tt.target.targetMean-podStatus.cpiMean) > epsilon || math.Abs(tt.target.targetStdDev-podStatus.stdDev) > epsilon {
				t.Errorf("expireAndUpdateThreshold threshold errors")
			}
			if !podStatus.isOnline {
				return
			}
			targetPodStatus.cpiSeries.expire(tt.expireNano)
			if tt.fields.isOnline && !dataSeriesEqual(targetPodStatus.cpiSeries, podStatus.cpiSeries) {
				t.Errorf("expireAndUpdateThreshold cpiSeries errors")
			}
		})
	}
}

// clone creates and returns a deep copy of the current podStatus instance, including its series and metrics.
func (pod *podStatus) clone() *podStatus {
	target := &podStatus{
		isOnline:       pod.isOnline,
		Hierarchy:      pod.Hierarchy,
		cpiSeries:      newDataSeries(),
		cpuUsageSeries: newDataSeries(),
		containers:     pod.containers,
	}
	if pod.cpiSeries != nil {
		for _, nano := range pod.cpiSeries.timeline {
			value, ok := pod.cpiSeries.data[nano]
			if ok {
				target.cpiSeries.add(value, nano)
			}
		}
	}
	if pod.cpuUsageSeries != nil {
		for _, nano := range pod.cpuUsageSeries.timeline {
			value, ok := pod.cpuUsageSeries.data[nano]
			if ok {
				target.cpuUsageSeries.add(value, nano)
			}
		}
	}
	target.count = pod.count
	target.cpiMean = pod.cpiMean
	target.stdDev = pod.stdDev
	return target
}

// TestPodStatusCheckOutlier tests checkOutlier function
func TestPodStatusCheckOutlier(t *testing.T) {
	_, err := perf.CgroupStat(cgroup.AbsoluteCgroupPath("perf_event", "", ""), time.Millisecond, cpiConf)
	if err != nil {
		return
	}
	type fields struct {
		isOnline       bool
		cpiSeries      *dataSeries
		cpuUsageSeries *dataSeries
		cpiMean        float64
		count          int64
		stdDev         float64
	}
	type args struct {
		now      time.Time
		duration time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "cpi is bigger than threshold three times",
			fields: fields{
				isOnline: true,
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 2.0, 11: 2.0,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 3.1, 10: 3.1, 11: 3.1,
					},
				},
				cpiMean: 2.0,
				count:   11,
				stdDev:  0.5,
			},
			args: args{
				now:      time.Unix(0, 11),
				duration: time.Nanosecond * 5,
			},
			want: true,
		}, {
			name: "cpi is bigger than threshold three times but cpuUsage < 0.25",
			fields: fields{
				isOnline: true,
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 0.20, 11: 2.0,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 3.1, 10: 3.1, 11: 3.1,
					},
				},
				cpiMean: 2.0,
				count:   11,
				stdDev:  0.5,
			},
			args: args{
				now:      time.Unix(0, 11),
				duration: time.Nanosecond * 5,
			},
			want: false,
		},
		{
			name: "cpi is bigger than threshold three times but cpiMean count <= 10",
			fields: fields{
				isOnline: true,
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 2.0, 10: 0.20, 11: 2.0,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.1, 8: 2.1, 9: 3.1, 10: 3.1, 11: 3.1,
					},
				},
				cpiMean: 2.0,
				count:   11,
				stdDev:  0.5,
			},
			args: args{
				now:      time.Unix(0, 11),
				duration: time.Nanosecond * 5,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podStatus := &podStatus{
				isOnline:       tt.fields.isOnline,
				cpiMean:        tt.fields.cpiMean,
				count:          tt.fields.count,
				stdDev:         tt.fields.stdDev,
				cpiSeries:      tt.fields.cpiSeries,
				cpuUsageSeries: tt.fields.cpuUsageSeries,
			}
			if got := podStatus.checkOutlier(tt.args.now, tt.args.duration); got != tt.want {
				t.Errorf("podStatus.checkOutlier() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPodStatusCheckAntagonists tests checkAntagonists function
func TestPodStatusCheckAntagonists(t *testing.T) {
	type fields struct {
		cpiSeries      *dataSeries
		cpuUsageSeries *dataSeries
		cpiMean        float64
		stdDev         float64
	}
	type args struct {
		now              time.Time
		window           time.Duration
		antagonistMetric float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "everything is ok",
			fields: fields{
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.2, 8: 2.5, 9: 3.2,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 3.1, 2: 3.5, 3: 2.8, 4: 2.9, 5: 3.1, 6: 3.1, 7: 2.8, 8: 3.1, 9: 3.1,
					},
				},
				cpiMean: 2.0,
				stdDev:  0.2,
			},
			args: args{
				now:              time.Unix(0, 11),
				window:           time.Nanosecond * 10,
				antagonistMetric: defaultAntagonistMetric,
			},
			want: true,
		},
		{
			name: "some cpi is filtered",
			fields: fields{
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 6: 2.0, 7: 2.2, 8: 2.5, 9: 3.2, 10: 3.3, 11: 3.5,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 11},
					data: map[int64]float64{
						1: 2.4, 2: 2.4, 3: 2.4, 4: 2.4, 5: 2.4, 6: 2.4, 7: 2.4, 8: 2.4, 11: 2.4,
					},
				},
				cpiMean: 2.0,
				stdDev:  0.2,
			},
			args: args{
				now:              time.Unix(0, 11),
				window:           time.Nanosecond * 10,
				antagonistMetric: defaultAntagonistMetric,
			},
			want: false,
		},
		{
			name: "some cpu usage is not collected",
			fields: fields{
				cpuUsageSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
					data: map[int64]float64{
						1: 2.0, 2: 2.1, 3: 2.1, 4: 2.0, 5: 2.1, 8: 2.5, 9: 3.2, 10: 3.3, 11: 3.5,
					},
				},
				cpiSeries: &dataSeries{
					timeline: []int64{1, 2, 3, 4, 5, 6, 7, 8, 11},
					data: map[int64]float64{
						1: 2.4, 2: 2.4, 3: 2.4, 4: 2.4, 5: 2.4, 6: 2.4, 7: 2.4, 8: 2.4, 11: 2.4,
					},
				},
				cpiMean: 2.0,
				stdDev:  0.2,
			},
			args: args{
				now:              time.Unix(0, 11),
				window:           time.Nanosecond * 10,
				antagonistMetric: defaultAntagonistMetric,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onlinePod := &podStatus{
				cpiSeries: tt.fields.cpiSeries,
				cpiMean:   tt.fields.cpiMean,
				stdDev:    tt.fields.stdDev,
			}
			offlinePod := &podStatus{
				cpuUsageSeries: tt.fields.cpuUsageSeries,
			}
			if got := onlinePod.checkAntagonists(tt.args.now, tt.args.window, offlinePod, tt.args.antagonistMetric); got != tt.want {
				t.Errorf("podStatus.isAntagonists() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPodStatusCollectData tests collectData function
func TestPodStatusCollectData(t *testing.T) {
	_, err := perf.CgroupStat(cgroup.AbsoluteCgroupPath("perf_event", "", ""), time.Millisecond, cpiConf)
	if err != nil {
		return
	}
	sampleDur := time.Second
	type fields struct {
		isOnline       bool
		Hierarchy      *cgroup.Hierarchy
		cpiSeries      *dataSeries
		cpuUsageSeries *dataSeries
	}
	type args struct {
		sampleDur time.Duration
		nowNano   int64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		dataSize int
	}{
		{
			name: "collect online pod",
			fields: fields{
				isOnline: true,
				Hierarchy: &cgroup.Hierarchy{
					Path: "",
				},
				cpiSeries:      newDataSeries(),
				cpuUsageSeries: newDataSeries(),
			},
			args: args{
				sampleDur: sampleDur,
				nowNano:   time.Now().UnixNano(),
			},
			dataSize: 1,
		},
		{
			name: "collect offline pod",
			fields: fields{
				isOnline: true,
				Hierarchy: &cgroup.Hierarchy{
					Path: "",
				},
				cpiSeries:      newDataSeries(),
				cpuUsageSeries: newDataSeries(),
			},
			args: args{
				sampleDur: sampleDur,
				nowNano:   time.Now().UnixNano(),
			},
			dataSize: 1,
		},
		{
			name: "cgroupPath is not exist ",
			fields: fields{
				isOnline: true,
				Hierarchy: &cgroup.Hierarchy{
					Path: "errorPath",
				},
				cpiSeries:      newDataSeries(),
				cpuUsageSeries: newDataSeries(),
			},
			args: args{
				sampleDur: sampleDur,
				nowNano:   time.Now().UnixNano(),
			},
			dataSize: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podStatus := &podStatus{
				isOnline:       tt.fields.isOnline,
				Hierarchy:      tt.fields.Hierarchy,
				cpiSeries:      tt.fields.cpiSeries,
				cpuUsageSeries: tt.fields.cpuUsageSeries,
			}
			podStatus.collectData(tt.args.sampleDur, tt.args.nowNano)
			time.Sleep(defaultSampleDur)
			if podStatus.isOnline && (len(podStatus.cpiSeries.data) != tt.dataSize || len(podStatus.cpuUsageSeries.data) != tt.dataSize) {
				t.Errorf("want dataSize = %v, got %v", tt.dataSize, len(podStatus.cpiSeries.data))
			}
			if !podStatus.isOnline && len(podStatus.cpuUsageSeries.data) != tt.dataSize {
				t.Errorf("want dataSize = %v, got %v", tt.dataSize, len(podStatus.cpiSeries.data))
			}
		})
	}
}

// TestPodStatusLimit tests limit function
func TestPodStatuslimit(t *testing.T) {
	_, err := perf.CgroupStat(cgroup.AbsoluteCgroupPath("perf_event", "", ""), time.Millisecond, cpiConf)
	if err != nil {
		return
	}
	type fields struct {
		Hierarchy   *cgroup.Hierarchy
		containers  map[string]*containerStatus
		isLimited   bool
		originQuota string
		podMutex    sync.RWMutex
	}
	type args struct {
		quota    string
		limitDur time.Duration
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		delayerIsNil bool
	}{
		{
			name: "pod is limited",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: "/",
					Path:       "tmp",
				},
				containers: map[string]*containerStatus{
					"container1": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: "/",
							Path:       "tmp/container1",
						},
						preCpuQuota: "10000",
					},
					"container2": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: "/",
							Path:       "tmp/container2",
						},
						preCpuQuota: "10000",
					},
				},
				isLimited:   true,
				originQuota: "10000",
			},
			args: args{
				quota:    "1000",
				limitDur: time.Second * 2,
			},
			wantErr:      false,
			delayerIsNil: false,
		},
		{
			name: "pod is not limited but delayer is not nil",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: "/",
					Path:       "tmp",
				},
				containers: map[string]*containerStatus{
					"container1": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: "/",
							Path:       "tmp/container1",
						},
						preCpuQuota: "10000",
					},
					"container2": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: "/",
							Path:       "tmp/container2",
						},
						preCpuQuota: "10000",
					},
				},
				isLimited:   false,
				originQuota: "10000",
			},
			args: args{
				quota:    "1000",
				limitDur: time.Second * 2,
			},
			wantErr:      false,
			delayerIsNil: false,
		},
		{
			name: "pod is not limited and delayer is nil",
			fields: fields{
				Hierarchy: &cgroup.Hierarchy{
					MountPoint: "/",
					Path:       "tmp",
				},
				containers: map[string]*containerStatus{
					"container1": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: "/",
							Path:       "tmp/container1",
						},
						preCpuQuota: "10000",
					},
					"container2": {
						Hierarchy: &cgroup.Hierarchy{
							MountPoint: "/",
							Path:       "tmp/container2",
						},
						preCpuQuota: "10000",
					},
				},
				isLimited:   false,
				originQuota: "10000",
			},
			args: args{
				quota:    "1000",
				limitDur: time.Second * 2,
			},
			wantErr:      false,
			delayerIsNil: true,
		},
	}
	for _, tt := range tests {
		createFile := func(dir string, fileName string, content string) {
			if err := os.MkdirAll(dir, constant.DefaultDirMode); err != nil {
				t.Errorf("error creating temp dir: %v", err)
			}
			file, err := os.Create(path.Join(dir, fileName))
			file.Chmod(constant.DefaultFileMode)
			if err != nil {
				t.Errorf("error creating quota file: %v", err)
			}
			if _, err := file.Write([]byte(content)); err != nil {
				t.Errorf("error writing to quota file: %v", err)
			}
			file.Close()
		}

		t.Run(tt.name, func(t *testing.T) {
			podDir := "/cpu/tmp"
			fileName := "cpu.cfs_quota_us"
			createFile(podDir, fileName, tt.fields.originQuota)
			containerDir1 := "/cpu/tmp/container1"
			containerDir2 := "/cpu/tmp/container2"
			createFile(containerDir1, fileName, tt.fields.originQuota)
			createFile(containerDir2, fileName, tt.fields.originQuota)
			defer os.RemoveAll(podDir)

			podStatus := &podStatus{
				Hierarchy:   tt.fields.Hierarchy,
				preCpuQuota: tt.fields.originQuota,
				podMutex:    tt.fields.podMutex,
				containers:  tt.fields.containers,
			}
			if !tt.delayerIsNil {
				podStatus.limit(tt.args.quota, tt.args.limitDur)
			}
			if !tt.fields.isLimited {
				time.Sleep(tt.args.limitDur)
			}
			err := podStatus.limit(tt.args.quota, tt.args.limitDur)
			if (err != nil) != tt.wantErr {
				t.Errorf("podStatus.limitCpuQuota() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				quota := podStatus.GetCgroupAttr(cpuQuotaKey).Value
				if quota != tt.args.quota {
					t.Errorf("limit pod quota errors")
				}
				for _, container := range podStatus.containers {
					containerQuota := container.GetCgroupAttr(cpuQuotaKey).Value
					if containerQuota != tt.args.quota {
						t.Errorf("limit container quota errors")
					}
				}

				time.Sleep(tt.args.limitDur + time.Second)

				quota = podStatus.GetCgroupAttr(cpuQuotaKey).Value
				if quota != tt.fields.originQuota {
					t.Errorf("recover pod quota errors")
				}
				for _, container := range podStatus.containers {
					containerQuota := container.GetCgroupAttr(cpuQuotaKey).Value
					if containerQuota != container.preCpuQuota {
						t.Errorf("limit container quota errors")
					}
				}

			}
		})
	}
}
