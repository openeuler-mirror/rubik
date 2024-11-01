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
// Description: This file defines the commonlib for resource manager

// Package common defines the commonlib for resource manager
package common

import (
	v2 "github.com/google/cadvisor/info/v2"
)

// PodStat is the status of pod
type PodStat struct {
	v2.ContainerInfo
}

// GetOption is the option to get podStat
type GetOption struct {
	CadvisorV2RequestOptions v2.RequestOptions
}

// Manager is the function set of Resource Manager
type Manager interface {
	Start() error
	Stop() error
	GetPodStats(name string, opt GetOption) (map[string]PodStat, error)
}
