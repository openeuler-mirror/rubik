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
// Description: memory status functions

package memory

import log "isula.org/rubik/pkg/tinylog"

const (
	// lowPressure means free / total < 30%
	lowPressure      = 0.3
	midPressure      = 0.15
	highPressure     = 0.1
	criticalPressure = 0.05
)

type levelInt int

const (
	normal levelInt = iota
	relieve
	low
	mid
	high
	critical
)

type status struct {
	pressureLevel levelInt
	relieveCnt    int
}

func newStatus() status {
	return status{
		pressureLevel: normal,
	}
}

func (s *status) set(pressureLevel levelInt) {
	s.pressureLevel = pressureLevel
	s.relieveCnt = 0
}

func (s *status) isNormal() bool {
	return s.pressureLevel == normal
}

func (s *status) isRelieve() bool {
	return s.pressureLevel == relieve
}

func (s *status) transitionStatus(freePercentage float64) {
	if freePercentage > lowPressure {
		switch s.pressureLevel {
		case normal:
		case low, mid, high, critical:
			log.Logf("change status from pressure to relieve")
			s.set(relieve)
		case relieve:
			if s.relieveCnt == relieveMaxCnt {
				s.set(normal)
				log.Logf("change status from relieve to normal")
			}
		}
		return
	}
	s.pressureLevel = getLevelInPressure(freePercentage)
}

func (s *status) String() string {
	switch s.pressureLevel {
	case normal:
		return "normal"
	case relieve:
		return "relieve"
	case low:
		return "low"
	case mid:
		return "mid"
	case high:
		return "high"
	case critical:
		return "critical"
	default:
		return "unknown"
	}
}

func getLevelInPressure(freePercentage float64) levelInt {
	var pressureLevel levelInt
	if freePercentage <= criticalPressure {
		pressureLevel = critical
	} else if freePercentage <= highPressure {
		pressureLevel = high
	} else if freePercentage <= midPressure {
		pressureLevel = mid
	} else {
		pressureLevel = low
	}
	return pressureLevel
}
