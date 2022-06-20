// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Yang Feiyu
// Create: 2022-6-7
// Description: tests for memory status functions

package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetStatus(t *testing.T) {
	s := newStatus()
	s.pressureLevel = relieve
	s.set(normal)
	assert.Equal(t, s.pressureLevel, normal)
	assert.Equal(t, s.relieveCnt, 0)
}

func TestIsNormal(t *testing.T) {
	s := newStatus()
	s.pressureLevel = relieve
	assert.Equal(t, s.isNormal(), false)
	s.set(normal)
	assert.Equal(t, s.isNormal(), true)
	assert.Equal(t, s.relieveCnt, 0)
}

func TestIsRelieve(t *testing.T) {
	s := newStatus()
	s.pressureLevel = relieve
	assert.Equal(t, s.isRelieve(), true)
}

func TestGetLevelInPressure(t *testing.T) {
	tests := []struct {
		freePercentage float64
		level          levelInt
	}{
		{
			freePercentage: 0.04,
			level:          critical,
		},
		{
			freePercentage: 0.09,
			level:          high,
		},
		{
			freePercentage: 0.13,
			level:          mid,
		},
		{
			freePercentage: 0.25,
			level:          low,
		},
	}

	for _, tt := range tests {
		tmp := getLevelInPressure(tt.freePercentage)
		assert.Equal(t, tmp, tt.level)
	}
}

func TestTransitionStatus(t *testing.T) {
	s := newStatus()
	s.transitionStatus(0.04)
	assert.Equal(t, s.pressureLevel, critical)

	s.transitionStatus(0.6)
	assert.Equal(t, s.pressureLevel, relieve)
	s.relieveCnt = relieveMaxCnt

	s.transitionStatus(0.6)
	assert.Equal(t, s.pressureLevel, normal)
}
