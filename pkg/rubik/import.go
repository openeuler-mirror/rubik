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
// Description: This file defines the services supported by rubik

// Package rubik defines the overall logic
package rubik

import (
	// introduce packages to auto register service
	_ "isula.org/rubik/pkg/services/dyncache"
	_ "isula.org/rubik/pkg/services/iocost"
	_ "isula.org/rubik/pkg/services/preemption"
	_ "isula.org/rubik/pkg/services/quotaburst"
	_ "isula.org/rubik/pkg/services/quotaturbo"
	_ "isula.org/rubik/pkg/version"
)
