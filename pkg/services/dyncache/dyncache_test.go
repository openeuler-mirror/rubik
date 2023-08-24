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
// Description: This file is testcase for cache limit service

// Package dyncache is the service used for cache limit setting
package dyncache

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/services/helper"
)

const (
	moduleName = "dynCache"
)

// TestCacheLimit_StartDynamic tests startDynamic of CacheLimit
func TestCacheLimit_Validate(t *testing.T) {
	const num2 = 2
	type fields struct {
		Config *Config
		Attr   *Attr
		Name   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantMsg string
	}{
		{
			name: "TC-static mode config",
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   minAdjustInterval + 1,
					PerfDuration:     minPerfDur + 1,
					L3Percent: MultiLvlPercent{
						Low:  minPercent + 1,
						Mid:  maxPercent/num2 + 1,
						High: maxPercent - 1,
					},
					MemBandPercent: MultiLvlPercent{
						Low:  minPercent + 1,
						Mid:  maxPercent/num2 + 1,
						High: maxPercent - 1,
					},
				},
			},
		},
		{
			name:    "TC-invalid mode config",
			wantErr: true,
			wantMsg: modeDynamic,
			fields: fields{
				Config: &Config{
					DefaultLimitMode: "invalid mode",
				},
			},
		},
		{
			name:    "TC-invalid adjust interval less than min value",
			wantErr: true,
			wantMsg: strconv.Itoa(minAdjustInterval),
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   minAdjustInterval - 1,
				},
			},
		},
		{
			name:    "TC-invalid adjust interval greater than max value",
			wantErr: true,
			wantMsg: strconv.Itoa(maxAdjustInterval),
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   maxAdjustInterval + 1,
				},
			},
		},
		{
			name:    "TC-invalid perf duration less than min value",
			wantErr: true,
			wantMsg: strconv.Itoa(minPercent),
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   maxAdjustInterval/num2 + 1,
					PerfDuration:     minPerfDur - 1,
				},
			},
		},
		{
			name:    "TC-invalid perf duration greater than max value",
			wantErr: true,
			wantMsg: strconv.Itoa(maxPerfDur),
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   maxAdjustInterval/num2 + 1,
					PerfDuration:     maxPerfDur + 1,
				},
			},
		},
		{
			name:    "TC-invalid percent value",
			wantErr: true,
			wantMsg: strconv.Itoa(minPercent),
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   maxAdjustInterval/num2 + 1,
					PerfDuration:     maxPerfDur/num2 + 1,
					L3Percent: MultiLvlPercent{
						Low: minPerfDur - 1,
					},
				},
			},
		},
		{
			name:    "TC-invalid l3 percent low value larger than mid value",
			wantErr: true,
			wantMsg: "low<=mid<=high",
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   maxAdjustInterval/num2 + 1,
					PerfDuration:     maxPerfDur/num2 + 1,
					L3Percent: MultiLvlPercent{
						Low:  minPercent + num2,
						Mid:  minPercent + 1,
						High: minPercent + 1,
					},
					MemBandPercent: MultiLvlPercent{
						Low:  minPercent,
						Mid:  minPercent + 1,
						High: minPercent + num2,
					},
				},
			},
		},
		{
			name:    "TC-invalid memband percent mid value larger than high value",
			wantErr: true,
			wantMsg: "low<=mid<=high",
			fields: fields{
				Config: &Config{
					DefaultLimitMode: modeStatic,
					AdjustInterval:   maxAdjustInterval/num2 + 1,
					PerfDuration:     maxPerfDur/num2 + 1,
					L3Percent: MultiLvlPercent{
						Low:  minPercent,
						Mid:  minPercent + 1,
						High: minPercent + num2,
					},
					MemBandPercent: MultiLvlPercent{
						Low:  minPercent,
						Mid:  maxPercent/num2 + 1,
						High: maxPercent / num2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DynCache{
				config: tt.fields.Config,
				Attr:   tt.fields.Attr,
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
			}
			err := c.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheLimit.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("CacheLimit.Validate() error = %v, wantMsg %v", err, tt.wantMsg)
			}
		})
	}
}

func TestNewCacheLimit(t *testing.T) {
	tests := []struct {
		name string
		want *DynCache
	}{
		{
			name: "TC-do nothing",
			want: &DynCache{
				ServiceBase: helper.ServiceBase{
					Name: moduleName,
				},
				Attr: &Attr{
					NumaNodeDir: defaultNumaNodeDir,
					MaxMiss:     defaultMaxMiss,
					MinMiss:     defaultMinMiss,
				},
				config: &Config{
					DefaultLimitMode:    modeStatic,
					DefaultResctrlDir:   defaultResctrlDir,
					DefaultPidNameSpace: defaultPidNameSpace,
					AdjustInterval:      defaultAdInt,
					PerfDuration:        defaultPerfDur,
					L3Percent: MultiLvlPercent{
						Low:  defaultLowL3,
						Mid:  defaultMidL3,
						High: defaultHighL3,
					},
					MemBandPercent: MultiLvlPercent{
						Low:  defaultLowMB,
						Mid:  defaultMidMB,
						High: defaultHighMB,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newDynCache(moduleName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCacheLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCacheLimit_PreStart tests PreStart
func TestCacheLimit_PreStart(t *testing.T) {
	type fields struct {
		Config *Config
		Attr   *Attr
		Viewer api.Viewer
		Name   string
	}
	type args struct {
		viewer api.Viewer
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		preHook  func(t *testing.T, c *DynCache)
		postHook func(t *testing.T, c *DynCache)
	}{
		{
			name:    "TC-just call function",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DynCache{
				config: tt.fields.Config,
				Attr:   tt.fields.Attr,
				Viewer: tt.fields.Viewer,
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
			}
			if err := c.PreStart(tt.args.viewer); (err != nil) != tt.wantErr {
				t.Errorf("CacheLimit.PreStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCacheLimit_ID tests ID
func TestCacheLimit_ID(t *testing.T) {
	type fields struct {
		Config *Config
		Attr   *Attr
		Viewer api.Viewer
		Name   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TC-return service's name",
			fields: fields{
				Name: "cacheLimit",
			},
			want: "cacheLimit",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DynCache{
				config: tt.fields.Config,
				Attr:   tt.fields.Attr,
				Viewer: tt.fields.Viewer,
				ServiceBase: helper.ServiceBase{
					Name: tt.fields.Name,
				},
			}
			if got := c.ID(); got != tt.want {
				t.Errorf("CacheLimit.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}
