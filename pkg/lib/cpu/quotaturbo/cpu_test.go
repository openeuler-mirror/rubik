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
// Date: 2023-02-20
// Description: This file is used for testing cpu.go

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculateUtils tests calculateUtils
func TestCalculateUtils(t *testing.T) {
	var (
		n1 float64 = 1
		n2 float64 = 2
		n3 float64 = 3
		n4 float64 = 4
	)

	var (
		t1 = ProcStat{
			total: n2,
			busy:  n1,
		}
		t2 = ProcStat{
			total: n4,
			busy:  n2,
		}
		t3 = ProcStat{
			total: n3,
			busy:  n3,
		}
	)
	// normal return result
	const (
		util               float64 = 50
		minimumUtilization float64 = 0
		maximumUtilization float64 = 100
	)
	assert.Equal(t, util, calculateUtils(t1, t2))
	// busy errors
	assert.Equal(t, minimumUtilization, calculateUtils(t2, t1))
	// total errors
	assert.Equal(t, maximumUtilization, calculateUtils(t2, t3))
}
