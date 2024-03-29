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
// Description: This file contains publisher factory

// Package publisher implement publisher interface
package publisher

import "isula.org/rubik/pkg/api"

type publihserType int8

const (
	// GENERIC indicates the generic publisher type
	GENERIC publihserType = iota
)

// Factory is the factory class of the publisher entity
type Factory struct {
}

var publisherFactory = &Factory{}

// GetPublisherFactory creates a publisher factory instance
func GetPublisherFactory() *Factory {
	return publisherFactory
}

// GetPublisher returns the publisher entity according to the publisher type
func (f *Factory) GetPublisher(publisherType publihserType) api.Publisher {
	switch publisherType {
	case GENERIC:
		return getGenericPublisher()
	default:
		return nil
	}
}
