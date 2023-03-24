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
// Date: 2023-03-24
// Description: This file is used for testing math

// Package util is common utilitization
package util

import (
	"testing"
)

// TestDiv tests Div
func TestDiv(t *testing.T) {
	const (
		dividend float64 = 100.0
		divisor  float64 = 1.0
		maxValue float64 = 70.0
		accuracy float64 = 2.0
		format   string  = "%.2f"
	)
	type args struct {
		dividend float64
		divisor  float64
		args     []interface{}
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "TC1-dividend: 100, divisor: 1, default arguments",
			args: args{
				dividend: dividend,
				divisor:  divisor,
			},
			want: dividend,
		},
		{
			name: "TC2-dividend: 100, divisor: 0, maxValue: 70",
			args: args{
				dividend: dividend,
				divisor:  0,
				args: []interface{}{
					maxValue,
				},
			},
			want: maxValue,
		},
		{
			name: "TC3-dividend: 100, divisor: 1, maxValue: 70, accuracy: 2",
			args: args{
				dividend: dividend,
				divisor:  divisor,
				args: []interface{}{
					maxValue,
					accuracy,
				},
			},
			want: maxValue,
		},
		{
			name: `TC4-dividend: 3, divisor: 8, format: %.2f`,
			args: args{
				dividend: 3,
				divisor:  8,
				args: []interface{}{
					maxValue,
					accuracy,
					format,
				},
			},
			want: 0.38,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Div(tt.args.dividend, tt.args.divisor, tt.args.args...); got != tt.want {
				t.Errorf("Div() = %v, want %v", got, tt.want)
			}
		})
	}
}
