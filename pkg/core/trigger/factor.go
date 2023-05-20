// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-05-19
// Description: This file is used for defining trigger factor

package trigger

import "isula.org/rubik/pkg/core/typedef"

// Factor represents the event trigger factor
type Factor interface {
	Message() string
	TargetPods() map[string]*typedef.PodInfo
}

// FactorImpl is the basic implementation of the trigger factor
type FactorImpl struct {
	Msg  string
	Pods map[string]*typedef.PodInfo
}

// Message returns the string information carried by Factor
func (impl *FactorImpl) Message() string {
	return impl.Msg
}

// TargetPods returns the pods that need to be processed
func (impl *FactorImpl) TargetPods() map[string]*typedef.PodInfo {
	return impl.Pods
}
