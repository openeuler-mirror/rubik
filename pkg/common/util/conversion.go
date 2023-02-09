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
// Date: 2023-02-08
// Description: This package stores basic type conversion functions.

// Package util is common utilitization
package util

import (
	"strconv"
)

// FormatInt64 convert the int 64 type to a string
func FormatInt64(n int64) string {
	const base = 10
	return strconv.FormatInt(n, base)
}

// ParseInt64 convert the string type to Int64
func ParseInt64(str string) (int64, error) {
	const (
		base    = 10
		bitSize = 64
	)
	return strconv.ParseInt(str, base, bitSize)
}

// ParseFloat64 convert the string type to Float64
func ParseFloat64(str string) (float64, error) {
	const bitSize = 64
	return strconv.ParseFloat(str, bitSize)
}

// FormatFloat64 convert the Float64 type to string
func FormatFloat64(f float64) string {
	const (
		precision = -1
		bitSize   = 64
		format    = 'f'
	)
	return strconv.FormatFloat(f, format, precision, bitSize)
}

// PercentageToDecimal converts percentages to decimals
func PercentageToDecimal(num float64) float64 {
	const percentageofOne float64 = 100
	return Div(num, percentageofOne)
}

// DeepCopy deep copy slice or map type data
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}
		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}
		return newSlice
	}
	return value
}
