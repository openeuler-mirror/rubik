// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-10-31
// Description: This file is used for maxValue transfomation

package executor

import (
	"context"
	"fmt"
	"math"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/trigger/common"
	"isula.org/rubik/pkg/core/trigger/template"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/resource/analyze"
)

// MaxValueTransformer returns a function that conforms to the Transformation format to filter for maximum utilization
func MaxValueTransformer(cal analyze.Calculator) template.Transformation {
	return func(ctx context.Context) (context.Context, error) {
		const epsilon = 1e-9
		var (
			chosen   *typedef.PodInfo
			maxValue float64 = -1
		)

		pods, ok := ctx.Value(common.TARGETPODS).(map[string]*typedef.PodInfo)
		if !ok {
			return ctx, fmt.Errorf("invalid target pod type")
		}

		for _, pod := range pods {
			value := cal(pod)
			// If the utilization rate is within the error, the two are considered to be the same.
			if maxValue > 0 && nearlyEqual(maxValue, value, epsilon) {
				if pod.StartTime.After(chosen.StartTime.Time) {
					maxValue = value
					chosen = pod
				}
				continue
			}
			if maxValue < value {
				maxValue = value
				chosen = pod
			}
		}
		// If the object is successfully obtained, the object is returned, otherwise an empty object is returned
		if chosen == nil {
			return context.WithValue(ctx, common.TARGETPODS, map[string]*typedef.PodInfo{}), nil
		}
		log.Infof("find the pod(%v) with the highest utilization(%.2f%%)", chosen.Name, maxValue)
		return context.WithValue(ctx, common.TARGETPODS, map[string]*typedef.PodInfo{chosen.Name: chosen}), nil
	}
}

func nearlyEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}
