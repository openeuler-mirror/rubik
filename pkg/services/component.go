// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2023-03-11
// Description: This file is used to initilize all components

// Package services
package services

import (
	"isula.org/rubik/pkg/feature"
	"isula.org/rubik/pkg/services/dyncache"
	"isula.org/rubik/pkg/services/dynmemory"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/pkg/services/iocost"
	"isula.org/rubik/pkg/services/iolimit"
	"isula.org/rubik/pkg/services/preemption"
	"isula.org/rubik/pkg/services/psi"
	"isula.org/rubik/pkg/services/quotaburst"
	"isula.org/rubik/pkg/services/quotaturbo"
)

// ServiceComponent is the handler function of initialization.
type ServiceComponent func(name string) error

var (
	serviceComponents = map[string]ServiceComponent{
		feature.PreemptionFeature: initPreemptionFactory,
		feature.DynCacheFeature:   initDynCacheFactory,
		feature.IOLimitFeature:    initIOLimitFactory,
		feature.IOCostFeature:     initIOCostFactory,
		feature.DynMemoryFeature:  initDynMemoryFactory,
		feature.QuotaBurstFeature: initQuotaBurstFactory,
		feature.QuotaTurboFeature: initQuotaTurboFactory,
		feature.PSIFeature:        initPSIFactory,
	}
)

func initIOLimitFactory(name string) error {
	return helper.AddFactory(name, iolimit.IOLimitFactory{ObjName: name})
}

func initIOCostFactory(name string) error {
	return helper.AddFactory(name, iocost.IOCostFactory{ObjName: name})
}

func initDynCacheFactory(name string) error {
	return helper.AddFactory(name, dyncache.DynCacheFactory{ObjName: name})
}

func initQuotaTurboFactory(name string) error {
	return helper.AddFactory(name, quotaturbo.QuotaTurboFactory{ObjName: name})
}

func initQuotaBurstFactory(name string) error {
	return helper.AddFactory(name, quotaburst.BurstFactory{ObjName: name})
}

func initPreemptionFactory(name string) error {
	return helper.AddFactory(name, preemption.PreemptionFactory{ObjName: name})
}

func initPSIFactory(name string) error {
	return helper.AddFactory(name, psi.Factory{ObjName: name})
}

func initDynMemoryFactory(name string) error {
	return helper.AddFactory(name, dynmemory.DynMemoryFactory{ObjName: name})
}
