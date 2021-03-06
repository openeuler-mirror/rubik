// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2022-03-25
// Description: This package stores basic type conversion functions.

package typedef

import (
	"crypto/rand"
	"math/big"
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

// RandInt provide safe rand int in range [0, max)
func RandInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}
