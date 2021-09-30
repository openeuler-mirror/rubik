// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2021-04-17
// Description: api definition for rubik

// Package api is for api definition
package api

// PodQoS describe Pod QoS settings
type PodQoS struct {
	CgroupPath string `json:"CgroupPath"`
	QosLevel   int    `json:"QosLevel"`
}

// SetQosRequest is request get from north end
type SetQosRequest struct {
	Pods map[string]PodQoS `json:"Pods"`
}

// SetQosResponse is response format for http responser
type SetQosResponse struct {
	ErrCode int    `json:"code"`
	Message string `json:"msg"`
}

// VersionResponse is version response for http responser
type VersionResponse struct {
	Version   string `json:"Version"`
	Release   string `json:"Release"`
	GitCommit string `json:"Commit"`
	BuildTime string `json:"BuildTime"`
	Usage     string `json:"Usage,omitempty"`
}
