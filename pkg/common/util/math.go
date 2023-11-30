// Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-02-08
// Description: This file is used for math

// Package util provide some util help functions.
package util

import (
	"fmt"
	"math"
)

// Div calculates the quotient of the divisor and the dividend, and it takes
// parameters (dividend, divisor, maximum out of range, precision, and format)
// format indicates the output format, for example, "%.2f" with two decimal places.
func Div(dividend, divisor float64, args ...interface{}) float64 {
	var (
		format   = ""
		accuracy = 1e-9
		maxValue = math.MaxFloat64
	)
	const (
		maxValueIndex int = iota
		accuracyIndex
		formatIndex
	)
	if len(args) > maxValueIndex {
		if value, ok := args[maxValueIndex].(float64); ok {
			maxValue = value
		}
	}
	if len(args) > accuracyIndex {
		if value, ok := args[accuracyIndex].(float64); ok {
			accuracy = value
		}
	}
	if len(args) > formatIndex {
		if value, ok := args[formatIndex].(string); ok {
			format = value
		}
	}

	if math.Abs(divisor) <= accuracy {
		return maxValue
	}
	ans := dividend / divisor
	if len(format) != 0 {
		if value, err := ParseFloat64(fmt.Sprintf(format, ans)); err == nil {
			ans = value
		}
	}
	return ans
}
