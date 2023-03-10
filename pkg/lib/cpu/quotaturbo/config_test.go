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

func TestConfig_SetAlarmWaterMark(t *testing.T) {
	type fields struct {
		HighWaterMark int
	}
	type args struct {
		arg int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TC1-set alarmWaterMark successfully",
			fields: fields{
				HighWaterMark: 60,
			},
			args: args{
				arg: 100,
			},
			wantErr: false,
		},
		{
			name: "TC2-alarmWaterMark = highwatermark",
			fields: fields{
				HighWaterMark: 60,
			},
			args: args{
				arg: 60,
			},
			wantErr: true,
		},
		{
			name: "TC2.1-alarmWaterMark < highwatermark",
			fields: fields{
				HighWaterMark: 60,
			},
			args: args{
				arg: 59,
			},
			wantErr: true,
		},
		{
			name: "TC3-alarmWaterMark > 100",
			fields: fields{
				HighWaterMark: 60,
			},
			args: args{
				arg: 101,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				HighWaterMark: tt.fields.HighWaterMark,
			}
			if err := c.SetAlarmWaterMark(tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("Config.SetAlarmWaterMark() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfig_SetHighWaterMark tests SetHighWaterMark of Config
func TestConfig_SetHighWaterMark(t *testing.T) {
	type fields struct {
		AlarmWaterMark int
	}
	type args struct {
		arg int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TC1-set highWaterMark successfully",
			fields: fields{
				AlarmWaterMark: 80,
			},
			args: args{
				arg: 10,
			},
			wantErr: false,
		},
		{
			name: "TC2-alarmWaterMark = highwatermark",
			fields: fields{
				AlarmWaterMark: 80,
			},
			args: args{
				arg: 80,
			},
			wantErr: true,
		},
		{
			name: "TC2.1-alarmWaterMark < highwatermark",
			fields: fields{
				AlarmWaterMark: 80,
			},
			args: args{
				arg: 81,
			},
			wantErr: true,
		},
		{
			name: "TC3-highWaterMark < 0",
			fields: fields{
				AlarmWaterMark: 60,
			},
			args: args{
				arg: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				AlarmWaterMark: tt.fields.AlarmWaterMark,
			}
			if err := c.SetHighWaterMark(tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("Config.SetHighWaterMark() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfig_SetlEvateLimit tests SetlEvateLimit of Config
func TestConfig_SetlEvateLimit(t *testing.T) {
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
			err := c.SetlEvateLimit(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.SetlEvateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				assert.Equal(t, tt.arg, c.ElevateLimit)
			} else {
				assert.Equal(t, defaultElevateLimit, c.ElevateLimit)
			}
		})
	}
}

// TestConfig_SetCPUFloatingLimit tests SetCPUFloatingLimit of Config
func TestConfig_SetCPUFloatingLimit(t *testing.T) {
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
			err := c.SetCPUFloatingLimit(tt.arg)
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

// TestOther tests other function of Config
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
			got.SetCgroupRoot(constant.TmpTestDir)
			assert.Equal(t, got.CgroupRoot, constant.TmpTestDir)
			got.SetSlowFallbackRatio(slowFallback)
			assert.Equal(t, got.SlowFallbackRatio, slowFallback)
			copyConf := got.GetConfig()
			if !reflect.DeepEqual(got, copyConf) {
				t.Errorf("GetConfig() = %v, want %v", got, copyConf)
			}
		})
	}
}
