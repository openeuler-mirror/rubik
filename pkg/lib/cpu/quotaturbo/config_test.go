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
// Date: 2023-03-07
// Description: This file is used for testing quota turbo config

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
)

// TestWithWaterMark tests WithWaterMark
func TestWithWaterMark(t *testing.T) {
	type fields struct {
		AlarmWaterMark int
		HighWaterMark  int
	}
	type args struct {
		highArg  int
		alarmArg int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TC1-set small WaterMark successfully",
			fields: fields{
				AlarmWaterMark: 80,
				HighWaterMark:  60,
			},
			args: args{
				highArg:  10,
				alarmArg: 20,
			},
			wantErr: false,
		},
		{
			name: "TC1.1-set large WaterMark successfully",
			fields: fields{
				AlarmWaterMark: 80,
				HighWaterMark:  60,
			},
			args: args{
				highArg:  85,
				alarmArg: 90,
			},
			wantErr: false,
		},
		{
			name: "TC2-alarmWaterMark = highwatermark",
			fields: fields{
				AlarmWaterMark: 80,
				HighWaterMark:  60,
			},
			args: args{
				highArg:  80,
				alarmArg: 80,
			},
			wantErr: true,
		},
		{
			name: "TC2.1-alarmWaterMark < highwatermark",
			fields: fields{
				AlarmWaterMark: 80,
				HighWaterMark:  60,
			},
			args: args{
				highArg:  81,
				alarmArg: 80,
			},
			wantErr: true,
		},
		{
			name: "TC3-highWaterMark < 0",
			fields: fields{
				AlarmWaterMark: 60,
				HighWaterMark:  30,
			},
			args: args{
				highArg:  -1,
				alarmArg: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				AlarmWaterMark: tt.fields.AlarmWaterMark,
				HighWaterMark:  tt.fields.AlarmWaterMark,
			}
			if err := WithWaterMark(tt.args.highArg, tt.args.alarmArg)(c); (err != nil) != tt.wantErr {
				t.Errorf("Config.SetHighWaterMark() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWithElevateLimit tests WithElevateLimit
func TestWithElevateLimit(t *testing.T) {
	const (
		normal   = 2.0
		larger   = 100.01
		negative = -0.01
	)
	tests := []struct {
		name    string
		arg     float64
		wantErr bool
	}{
		{
			name:    "TC1-set EvateLimit successfully",
			arg:     normal,
			wantErr: false,
		},
		{
			name:    "TC2-too large EvateLimit",
			arg:     larger,
			wantErr: true,
		},
		{
			name:    "TC3-negative EvateLimit",
			arg:     negative,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			err := WithElevateLimit(tt.arg)(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithElevateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				assert.Equal(t, tt.arg, c.ElevateLimit)
			} else {
				assert.Equal(t, defaultElevateLimit, c.ElevateLimit)
			}
		})
	}
}

// TestWithCPUFloatingLimit tests WithCPUFloatingLimit
func TestWithCPUFloatingLimit(t *testing.T) {
	const (
		normal   = 20.0
		larger   = 100.01
		negative = -0.01
	)
	tests := []struct {
		name    string
		arg     float64
		wantErr bool
	}{
		{
			name:    "TC1-set CPUFloatingLimit successfully",
			arg:     normal,
			wantErr: false,
		},
		{
			name:    "TC2-too large CPUFloatingLimit",
			arg:     larger,
			wantErr: true,
		},
		{
			name:    "TC3-negative CPUFloatingLimit",
			arg:     negative,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			err := WithCPUFloatingLimit(tt.arg)(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.SetCPUFloatingLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				assert.Equal(t, tt.arg, c.CPUFloatingLimit)
			} else {
				assert.Equal(t, defaultCPUFloatingLimit, c.CPUFloatingLimit)
			}
		})
	}
}

// TestOther tests other function
func TestOther(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "TC1-test other",
			want: &Config{
				HighWaterMark:     defaultHighWaterMark,
				AlarmWaterMark:    defaultAlarmWaterMark,
				CgroupRoot:        constant.DefaultCgroupRoot,
				ElevateLimit:      defaultElevateLimit,
				SlowFallbackRatio: defaultSlowFallbackRatio,
				CPUFloatingLimit:  defaultCPUFloatingLimit,
			},
		},
	}
	const slowFallback = 3.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
			WithCgroupRoot(constant.TmpTestDir)(got)
			assert.Equal(t, got.CgroupRoot, constant.TmpTestDir)
			WithSlowFallbackRatio(slowFallback)(got)
			assert.Equal(t, got.SlowFallbackRatio, slowFallback)
		})
	}
}
