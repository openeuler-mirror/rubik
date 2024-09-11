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
// Create: 2023-01-28
// Description: This file defines informerFactory which return the informer creator

// Package informer implements informer interface
package informer

import (
	"fmt"

	"isula.org/rubik/pkg/api"
)

type (
	// informer's factory class
	informerFactory struct{}
	informerCreator func(publisher api.Publisher) (api.Informer, error)
)

const (
	APISERVER = "apiserver" // the informer to interact with the apiserver of kubernetes
	NRI       = "nri"       // the informer to interact with the NRI interface
)

// defaultInformerFactory is globally unique informer factory
var defaultInformerFactory *informerFactory

// GetInformerCreator returns the constructor of the informer of the specified type
func (factory *informerFactory) GetInformerCreator(informerName string) informerCreator {
	switch informerName {
	case APISERVER:
		return NewAPIServerInformer
	case NRI:
		return NewNRIInformer
	default:
		return func(publisher api.Publisher) (api.Informer, error) {
			return nil, fmt.Errorf("informer not implemented")
		}
	}
}

// GetInformerFactory returns the Informer factory class entity
func GetInformerFactory() *informerFactory {
	if defaultInformerFactory == nil {
		defaultInformerFactory = &informerFactory{}
	}
	return defaultInformerFactory
}
