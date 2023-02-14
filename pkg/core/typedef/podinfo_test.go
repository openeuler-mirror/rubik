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
// Description: This file tests podInfo

// Package typedef defines core struct and methods for rubik
package typedef

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/common/constant"
)

func TestPodInfo_DeepCopy(t *testing.T) {
	const (
		oldPodName          = "FooPod"
		newPodName          = "NewFooPod"
		oldPodID            = "testPod1"
		newPodID            = "newTestPod1"
		oldQuota            = "true"
		newQuota            = "false"
		oldContName         = "FooCon"
		newContName         = "NewFooPod"
		oldReqCPU   float64 = 1.2
		newReqCPU   float64 = 2.7
		oldReqMem   float64 = 500
		newReqMem   float64 = 350
		contID              = "testCon1"
		oldLimitCPU float64 = 9.0
		oldLimitMem float64 = 300
	)
	oldPod := &PodInfo{
		Name: oldPodName,
		UID:  oldPodID,
		Annotations: map[string]string{
			constant.QuotaAnnotationKey:    oldQuota,
			constant.PriorityAnnotationKey: "true",
		},
		IDContainersMap: map[string]*ContainerInfo{
			contID: {
				Name:             oldContName,
				RequestResources: ResourceMap{ResourceCPU: oldReqCPU, ResourceMem: oldReqMem},
				LimitResources:   ResourceMap{ResourceCPU: 9.0, ResourceMem: 300},
			},
		},
	}
	copyPod := oldPod.DeepCopy()
	copyPod.Name = newContName
	copyPod.UID = newPodID
	copyPod.Annotations[constant.QuotaAnnotationKey] = newQuota
	copyPod.IDContainersMap[contID].Name = newContName
	copyPod.IDContainersMap[contID].RequestResources[ResourceCPU] = newReqCPU
	copyPod.IDContainersMap[contID].RequestResources[ResourceMem] = newReqMem

	assert.Equal(t, oldPodName, oldPod.Name)
	assert.Equal(t, oldPodID, oldPod.UID)
	assert.Equal(t, oldContName, oldPod.IDContainersMap[contID].Name)
	assert.Equal(t, oldQuota, oldPod.Annotations[constant.QuotaAnnotationKey])
	assert.Equal(t, oldReqCPU, oldPod.IDContainersMap[contID].RequestResources[ResourceCPU])
	assert.Equal(t, oldReqMem, oldPod.IDContainersMap[contID].RequestResources[ResourceMem])
	assert.Equal(t, oldLimitCPU, oldPod.IDContainersMap[contID].LimitResources[ResourceCPU])
	assert.Equal(t, oldLimitMem, oldPod.IDContainersMap[contID].LimitResources[ResourceMem])

}
