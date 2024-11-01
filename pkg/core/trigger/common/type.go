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
// Description: This file is used for defining trigger factor

package common

import "context"

type (
	// Factor is used to trigger various events
	Factor uint8
)

const (
	TARGETPODS Factor = iota
	DEPORTPOD
)

// Descriptor defines methods for describing triggers
type Descriptor interface {
	Name() string
}

// Executor is the trigger execution function interface
type Executor interface {
	Execute(context.Context) (context.Context, error)
	Stop() error
}

type Setter interface {
	SetNext(...Trigger) Trigger
}

// Trigger interface defines the trigger methods
type Trigger interface {
	Descriptor
	Setter
	Activate(context.Context) error
	Halt() error
}
