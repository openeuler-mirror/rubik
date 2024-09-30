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
// Description: This file for defining rubik support features

// Package rubik provide rubik main logic.
package rubik

import (
	"isula.org/rubik/pkg/feature"
	"isula.org/rubik/pkg/services"
)

var defaultRubikFeature = []services.FeatureSpec{
	{
		Name:    feature.PreemptionFeature,
		Default: true,
	},
	{
		Name:    feature.DynCacheFeature,
		Default: true,
	},
	{
		Name:    feature.IOLimitFeature,
		Default: true,
	},
	{
		Name:    feature.IOCostFeature,
		Default: true,
	},
	{
		Name:    feature.DynMemoryFeature,
		Default: true,
	},
	{
		Name:    feature.QuotaBurstFeature,
		Default: true,
	},
	{
		Name:    feature.QuotaTurboFeature,
		Default: true,
	},
	{
		Name:    feature.PSIFeature,
		Default: true,
	},
	{
		Name:    feature.CPIFeature,
		Default: true,
	},
}
