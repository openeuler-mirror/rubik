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
// Create: 2023-03-23
// Description: This file test common function of cgroup

// Package cgroup uses map to manage cgroup parameters and provides a friendly and simple cgroup usage method
package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
)

// TestReadCgroupFile tests ReadCgroupFile
func TestReadCgroupFile(t *testing.T) {
	InitMountDir(constant.TmpTestDir)
	defer InitMountDir(constant.DefaultCgroupRoot)
	pathElems := []string{"cpu", "kubepods", "PodXXX", "ContYYY", "cpu.cfs_quota_us"}
	const value = "-1\n"
	tests := []struct {
		name    string
		args    []string
		pre     func(t *testing.T)
		post    func(t *testing.T)
		want    []byte
		wantErr bool
	}{
		{
			name:    "TC1-non existed path",
			args:    pathElems,
			wantErr: true,
			want:    nil,
		},
		{
			name: "TC2-successfully",
			args: pathElems,
			pre: func(t *testing.T) {
				assert.NoError(t, util.WriteFile(
					filepath.Join(constant.TmpTestDir, filepath.Join(pathElems...)),
					value))
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: false,
			want:    []byte(value),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre(t)
			}
			got, err := ReadCgroupFile(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCgroupFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadCgroupFile() = %v, want %v", got, tt.want)
			}
			if tt.post != nil {
				tt.post(t)
			}
		})
	}
}

// TestWriteCgroupFile tests WriteCgroupFile
func TestWriteCgroupFile(t *testing.T) {
	InitMountDir(constant.TmpTestDir)
	defer InitMountDir(constant.DefaultCgroupRoot)
	pathElems := []string{"cpu", "kubepods", "PodXXX", "ContYYY", "cpu.cfs_quota_us"}
	const value = "-1\n"
	type args struct {
		content string
		elem    []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		pre     func(t *testing.T)
		post    func(t *testing.T)
	}{
		{
			name: "TC1-non existed path",
			args: args{
				content: value,
				elem:    pathElems,
			},
			wantErr: true,
		},
		{
			name: "TC2-successfully",
			args: args{
				content: value,
				elem:    pathElems,
			},
			pre: func(t *testing.T) {
				assert.NoError(t, util.WriteFile(
					filepath.Join(constant.TmpTestDir, filepath.Join(pathElems...)),
					value))
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre(t)
			}
			if err := WriteCgroupFile(tt.args.content, tt.args.elem...); (err != nil) != tt.wantErr {
				t.Errorf("WriteCgroupFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.post != nil {
				tt.post(t)
			}
		})
	}
}

// TestAttr_Expect tests Expect of Attr
func TestAttr_Expect(t *testing.T) {
	const (
		intValue     int     = 1
		int64Value   int64   = 1
		stringValue  string  = "rubik"
		float64Value float64 = 1.0
	)
	type fields struct {
		Value string
		Err   error
	}
	type args struct {
		arg interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TC1-attribute error",
			fields: fields{
				Value: "",
				Err:   fmt.Errorf("failed to get value"),
			},
			wantErr: true,
		},
		{
			name: "TC2.1-expect int 1(parse fail)",
			fields: fields{
				Value: "",
			},
			args: args{
				arg: intValue,
			},
			wantErr: true,
		},
		{
			name: "TC2.2-expect int 1(not equal value)",
			fields: fields{
				Value: "100",
			},
			args: args{
				arg: intValue,
			},
			wantErr: true,
		},
		{
			name: "TC2.3-expect int 1(success)",
			fields: fields{
				Value: "1",
			},
			args: args{
				arg: intValue,
			},
			wantErr: false,
		},
		{
			name: "TC3.1-expect int64 1(parse fail)",
			fields: fields{
				Value: "",
			},
			args: args{
				arg: int64Value,
			},
			wantErr: true,
		},
		{
			name: "TC3.2-expect int64 1(not equal value)",
			fields: fields{
				Value: "100",
			},
			args: args{
				arg: int64Value,
			},
			wantErr: true,
		},
		{
			name: "TC3.3-expect int64 1(success)",
			fields: fields{
				Value: "1",
			},
			args: args{
				arg: int64Value,
			},
			wantErr: false,
		},
		{
			name: "TC4.1-expect string rubik(not equal value)",
			fields: fields{
				Value: "-1",
			},
			args: args{
				arg: stringValue,
			},
			wantErr: true,
		},
		{
			name: "TC4.2-expect string rubik(success)",
			fields: fields{
				Value: stringValue,
			},
			args: args{
				arg: stringValue,
			},
			wantErr: false,
		},
		{
			name: "TC5-expect float64(undefined type)",
			args: args{
				arg: float64Value,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attr{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			if err := attr.Expect(tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("Attr.Expect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAttr_Int64 tests Int64 of Attr
func TestAttr_Int64(t *testing.T) {
	const int64Value int64 = 1
	type fields struct {
		Value string
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name: "TC1-attribute error",
			fields: fields{
				Err: fmt.Errorf("failed to get value"),
			},
			wantErr: true,
		},
		{
			name: "TC2-expect int64 1(success)",
			fields: fields{
				Value: "1",
			},
			want:    int64Value,
			wantErr: false,
		},
		{
			name: "TC3-expect int64 1(error parse)",
			fields: fields{
				Value: "rubik",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attr{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := attr.Int64()
			if (err != nil) != tt.wantErr {
				t.Errorf("Attr.Int64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Attr.Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAttr_Int tests Int of Attr
func TestAttr_Int(t *testing.T) {
	const intValue int = 1
	type fields struct {
		Value string
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			name: "TC1-attribute error",
			fields: fields{
				Err: fmt.Errorf("failed to get value"),
			},
			wantErr: true,
		},
		{
			name: "TC2-expect int 1(success)",
			fields: fields{
				Value: "1",
			},
			want:    intValue,
			wantErr: false,
		},
		{
			name: "TC3-expect int 1(error parse)",
			fields: fields{
				Value: "rubik",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attr{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := attr.Int()
			if (err != nil) != tt.wantErr {
				t.Errorf("Attr.Int() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Attr.Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAttr_Int64Map tests Int64Map of Attr
func TestAttr_Int64Map(t *testing.T) {
	m := map[string]int64{"a": 1, "b": 2}
	type fields struct {
		Value string
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]int64
		wantErr bool
	}{

		{
			name: "TC1-attribute error",
			fields: fields{
				Err: fmt.Errorf("failed to get value"),
			},
			wantErr: true,
		},
		{
			name: "TC2-expect int 1(success)",
			fields: fields{
				Value: `a 1
				b 2`,
			},
			want:    m,
			wantErr: false,
		},
		{
			name: "TC3-expect int 1(error parse)",
			fields: fields{
				Value: "rubik",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attr{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := attr.Int64Map()
			if (err != nil) != tt.wantErr {
				t.Errorf("Attr.Int64Map() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attr.Int64Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAttr_CPUStat tests CPUStat of Attr
func TestAttr_CPUStat(t *testing.T) {
	res := &CPUStat{
		NrPeriods:     1,
		NrThrottled:   1,
		ThrottledTime: 1,
	}
	type fields struct {
		Value string
		Err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    *CPUStat
		wantErr bool
	}{
		{
			name: "TC1-attribute error",
			fields: fields{
				Err: fmt.Errorf("failed to get value"),
			},
			wantErr: true,
		},
		{
			name: "TC2-expect int 1(success)",
			fields: fields{
				Value: `nr_periods 1
				nr_throttled 1
				throttled_time 1`,
			},
			want:    res,
			wantErr: false,
		},
		{
			name: "TC3-expect int 1(error parse)",
			fields: fields{
				Value: "rubik",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attr{
				Value: tt.fields.Value,
				Err:   tt.fields.Err,
			}
			got, err := attr.CPUStat()
			if (err != nil) != tt.wantErr {
				t.Errorf("Attr.CPUStat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attr.CPUStat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHierarchy_SetCgroupAttr tests SetCgroupAttr of Hierarchy
func TestHierarchy_SetCgroupAttr(t *testing.T) {
	type args struct {
		key   *Key
		value string
	}
	tests := []struct {
		name    string
		path    string
		args    args
		wantErr bool
	}{
		{
			name:    "TC1.1-empty key",
			args:    args{},
			wantErr: true,
		},
		{
			name: "TC1.2-empty Subsys",
			args: args{
				key: &Key{},
			},
			wantErr: true,
		},
		{
			name: "TC2-",
			args: args{
				key: &Key{
					SubSys:   "cpu",
					FileName: "cpu.cfs_quota_us",
				},
				value: "1",
			},
			path:    "kubepods/PodXXX/ContXXX",
			wantErr: true,
		},
	}
	defer os.RemoveAll(constant.TmpTestDir)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHierarchy(constant.TmpTestDir, tt.path)
			if err := h.SetCgroupAttr(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Hierarchy.SetCgroupAttr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHierarchy_GetCgroupAttr tests GetCgroupAttr of Hierarchy
func TestHierarchy_GetCgroupAttr(t *testing.T) {
	const (
		contPath = "kubepods/PodXXX/ContXXX"
		value    = " 1\n"
	)
	var quotaKey = &Key{
		SubSys:   "cpu",
		FileName: "cpu.cfs_quota_us",
	}
	tests := []struct {
		name    string
		path    string
		args    *Key
		want    string
		wantErr bool
		pre     func(t *testing.T)
		post    func(t *testing.T)
	}{
		{
			name:    "TC1.1-empty key",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "TC2-empty path",
			args:    quotaKey,
			path:    contPath,
			wantErr: true,
		},
		{
			name: "TC3-success",
			args: quotaKey,
			path: contPath,
			pre: func(t *testing.T) {
				assert.NoError(t, util.WriteFile(
					filepath.Join(constant.TmpTestDir, quotaKey.SubSys, contPath, quotaKey.FileName),
					value))
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			want:    "1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHierarchy(constant.TmpTestDir, tt.path)
			if tt.pre != nil {
				tt.pre(t)
			}
			got := h.GetCgroupAttr(tt.args)
			if (got.Err != nil) != tt.wantErr {
				t.Errorf("Hierarchy.GetCgroupAttr() error = %v, wantErr %v", got.Err, tt.wantErr)
			}
			if got.Err == nil {
				if got.Value != tt.want {
					t.Errorf("Hierarchy.GetCgroupAttr() = %v, want %v", got, tt.want)
				}
			}
			if tt.post != nil {
				tt.post(t)
			}
		})
	}
}

// TestAbsoluteCgroupPath tests AbsoluteCgroupPath
func TestAbsoluteCgroupPath(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "TC1-AbsoluteCgroupPath",
			args: []string{"a", "b"},
			want: GetMountDir() + "/a/b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AbsoluteCgroupPath(tt.args...); got != tt.want {
				t.Errorf("AbsoluteCgroupPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
