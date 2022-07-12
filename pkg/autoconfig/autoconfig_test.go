// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2022-07-10
// Description: autoconfig test

package autoconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// TestInit test Init
func TestInit(t *testing.T) {
	kb := &kubernetes.Clientset{}
	err := Init(kb)
	assert.Equal(t, true, strings.Contains(err.Error(), "environment variable"))
}

// TestAddUpdateDelHandler test Handler
func TestAddUpdateDelHandler(t *testing.T) {
	oldObj := corev1.Pod{}
	newObj := corev1.Pod{}
	addHandler(oldObj)
	updateHandler(oldObj, newObj)
	deleteHandler(oldObj)
}
