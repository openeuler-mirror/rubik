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
	"bufio"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

// ParseInt64Map converts string to map[string]Int64:
// 1. multiple lines are allowed;
// 2. a single line consists of only two strings separated by spaces
func ParseInt64Map(data string) (map[string]int64, error) {
	var res = make(map[string]int64)
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		arr := strings.Fields(scanner.Text())
		const defaultLength = 2
		if len(arr) != defaultLength {
			return nil, fmt.Errorf("fail to parse a single line into two strings %v", arr)
		}
		value, err := ParseInt64(arr[1])
		if err != nil {
			return nil, err
		}
		res[arr[0]] = value
	}
	return res, nil
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

// DeepCopy deep copy slice or map data with basic type
// DeepCopy is a simple deep copy function that only supports copying of basic types of maps and slices,
// such as map[string]string, []int, etc., and does not support nesting.
// **Since the reflection mechanism with poor performance is used,
// please use this function with caution after balancing performance and ease of use**
func DeepCopy(value interface{}) interface{} {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}()
	typ := reflect.TypeOf(value)
	val := reflect.ValueOf(value)
	switch typ.Kind() {
	case reflect.Map:
		newMap := reflect.MakeMapWithSize(typ, val.Len())
		it := val.MapRange()
		for it.Next() {
			newMap.SetMapIndex(it.Key(), DeepCopy(it.Value()).(reflect.Value))
		}
		return newMap.Interface()
	case reflect.Slice:
		newSlice := reflect.MakeSlice(typ, val.Len(), val.Cap())
		reflect.Copy(newSlice, val)
		return newSlice.Interface()
	}
	return value
}
