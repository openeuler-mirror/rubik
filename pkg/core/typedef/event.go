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
	// EventType is the type of event published by generic publisher
	EventType int8
	// Event is the event published by generic publisher
	Event interface{}
)

const (
	// RAWPODADD means Kubernetes starts a new Pod event
	RAWPODADD EventType = iota
	// RAWPODUPDATE means Kubernetes updates Pod event
	RAWPODUPDATE
	// RAWPODDELETE means Kubernetes deletes Pod event
	RAWPODDELETE
	// INFOADD means PodManager adds pod information event
	INFOADD
	// INFOUPDATE means PodManager updates pod information event
	INFOUPDATE
	// INFODELETE means PodManager deletes pod information event
	INFODELETE
	// RAWPODSYNCALL means Full amount of kubernetes pods
	RAWPODSYNCALL
	// NRIPODADD means nri starts a new Pod event
	NRIPODADD
	// NRIPODDELETE means nri deletes Pod event
	NRIPODDELETE
	// NRICONTAINERCREATE means nri starts container event
	NRICONTAINERSTART
	// NRICONTAINERREMOVE means nri removes Container event
	NRICONTAINERREMOVE
	// NRIPODSYNCALL means sync all Pods event
	NRIPODSYNCALL
	// NRICONTAINERSYNCALL means sync all Containers event
	NRICONTAINERSYNCALL
)

const undefinedType = "undefined"

var eventTypeToString = map[EventType]string{
	RAWPODADD:           "addrawpod",
	RAWPODUPDATE:        "updaterawpod",
	RAWPODDELETE:        "deleterawpod",
	INFOADD:             "addinfo",
	INFOUPDATE:          "updateinfo",
	INFODELETE:          "deleteinfo",
	RAWPODSYNCALL:       "syncallrawpods",
	NRIPODADD:           "addnripod",
	NRIPODDELETE:        "deletenripod",
	NRICONTAINERSTART:   "createnricontainer",
	NRICONTAINERREMOVE:  "removenricontainer",
	NRIPODSYNCALL:       "syncallnrirawpods",
	NRICONTAINERSYNCALL: "syncallnrirawcontainers",
}

// String returns the string of the current event type
func (t EventType) String() string {
	if str, ok := eventTypeToString[t]; ok {
		return str
	}
	return undefinedType
}
