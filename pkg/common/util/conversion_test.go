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
// Date: 2023-02-14
// Description: This file is used for testing conversion functions

// Package util is common utilitization
package util

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormatInt64 is testcase for FormatInt64
func TestFormatInt64(t *testing.T) {
	type args struct {
		n int64
	}
	validNum := 100
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TC-convert the int 64 to string",
			args: args{n: int64(validNum)},
			want: "100",
		},
		{
			name: "TC-convert the big int",
			args: args{math.MaxInt64},
			want: "9223372036854775807",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatInt64(tt.args.n); got != tt.want {
				t.Errorf("FormatInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseInt64 is testcase for ParseInt64
func TestParseInt64(t *testing.T) {
	type args struct {
		str string
	}
	validNum := 100
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "TC-convert the int 64 to string",
			args: args{str: "100"},
			want: int64(validNum),
		},
		{
			name: "TC-convert the big int",
			args: args{str: "9223372036854775807"},
			want: math.MaxInt64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt64(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatFloat64 is testcase for FormatFloat64
func TestFormatFloat64(t *testing.T) {
	type args struct {
		f float64
	}
	validNum := 100.0
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TC-convert the float64 to string",
			args: args{f: validNum},
			want: "100",
		},
		{
			name: "TC-convert the big float",
			args: args{math.MaxFloat64},
			want: "179769313486231570000000000000000000000000000000000000000000000000000000000000000000" +
				"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
				"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
				"00000000000000000000000000000000000000000000000000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFloat64(tt.args.f); got != tt.want {
				t.Errorf("FormatFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseFloat64 is testcase for ParseFloat64
func TestParseFloat64(t *testing.T) {
	type args struct {
		str string
	}
	validNum := 100.0
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "TC-convert the string to float64",
			args: args{str: "100"},
			want: validNum,
		},
		{
			name: "TC-convert the big float",
			args: args{str: "1797693134862315700000000000000000000000000000000000000000000000000000000000" +
				"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
				"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
				"000000000000000000000000000000000000000000000000000000000"},
			want: math.MaxFloat64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloat64(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeepCopy tests DeepCopy
func TestDeepCopy(t *testing.T) {
	oldMap := map[string]string{
		"a": "1",
		"b": "2",
	}
	newMap := DeepCopy(oldMap).(map[string]string)
	newMap["a"] = "3"
	newMap["b"] = "4"
	assert.Equal(t, oldMap["a"], "1")
	assert.Equal(t, oldMap["b"], "2")
	assert.Equal(t, newMap["a"], "3")
	assert.Equal(t, newMap["b"], "4")

	oldSlice := []string{"a", "b", "c"}
	newSlice := DeepCopy(oldSlice).([]string)
	for i := range newSlice {
		newSlice[i] += "z"
	}
	assert.Equal(t, oldSlice, []string{"a", "b", "c"})
	assert.Equal(t, newSlice, []string{"az", "bz", "cz"})
}

// TestParseInt64Map tests ParseInt64Map
func TestParseInt64Map(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]int64
		wantErr bool
	}{
		{
			name: "TC1-length of fields is 3",
			args: args{
				data: `a 10 10`,
			},
			wantErr: true,
		},
		{
			name: "TC2-the second field is string type",
			args: args{
				data: `a a`,
			},
			wantErr: true,
		},
		{
			name: "TC3-success",
			args: args{
				data: `a 10`,
			},
			want:    map[string]int64{"a": 10},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt64Map(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInt64Map() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInt64Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPercentageToDecimal tests PercentageToDecimal
func TestPercentageToDecimal(t *testing.T) {
	type args struct {
		num float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "TC1-1% to 0.01",
			args: args{
				num: 1.0,
			},
			want: 0.01,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PercentageToDecimal(tt.args.num); got != tt.want {
				t.Errorf("PercentageToDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}
