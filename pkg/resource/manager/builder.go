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
// Description: This file defines the builder of resource managers

// Package manager implements manager builer
package manager

import (
	"fmt"

	"isula.org/rubik/pkg/resource/manager/cadvisor"
	"isula.org/rubik/pkg/resource/manager/common"
)

type managerTyp uint8

const (
	CADVISOR managerTyp = iota
)

type CadvisorConfig interface {
	Config() *cadvisor.Config
}

type Builder func(interface{}) (common.Manager, error)

func GetManagerBuilder(typ managerTyp) Builder {
	switch typ {
	case CADVISOR:
		return newCasvisorManagerBuilder()
	}
	return nil
}

func newCasvisorManagerBuilder() Builder {
	return func(args interface{}) (common.Manager, error) {
		conf, ok := args.(CadvisorConfig)
		if !ok {
			return nil, fmt.Errorf("failed to get cadvisor config")
		}
		return cadvisor.New(conf.Config()), nil
	}
}
