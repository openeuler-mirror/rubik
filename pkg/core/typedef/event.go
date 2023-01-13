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
// Create: 2023-01-05
// Description: This file defines event type and event

// Package typedef defines core struct and methods for rubik
package typedef

type (
	// the type of event published by generic publisher
	EventType int8
	// the event published by generic publisher
	Event interface{}
)

const (
	// Kubernetes starts a new Pod event
	RAW_POD_ADD EventType = iota
	// Kubernetes updates Pod event
	RAW_POD_UPDATE
	// Kubernetes deletes Pod event
	RAW_POD_DELETE
	// PodManager adds pod information event
	INFO_ADD
	// PodManager updates pod information event
	INFO_UPDATE
	// PodManager deletes pod information event
	INFO_DELETE
)

const unknownType = "unknown"

var eventTypeToString = map[EventType]string{
	RAW_POD_ADD:    "addrawpod",
	RAW_POD_UPDATE: "updaterawpod",
	RAW_POD_DELETE: "deleterawpod",
	INFO_ADD:       "addinfo",
	INFO_UPDATE:    "updateinfo",
	INFO_DELETE:    "deleteinfo",
}

// String returns the string of the current event type
func (t EventType) String() string {
	if str, ok := eventTypeToString[t]; ok {
		return str
	}
	return unknownType
}
