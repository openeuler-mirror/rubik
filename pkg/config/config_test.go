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
// Create: 2023-02-01
// Description: This file is used to test the functions of config.go

// Package config is used to manage the configuration of rubik
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
)

var rubikConfig string = `
{
	"agent": {
	  "logDriver": "stdio",
	  "logDir": "/var/log/rubik",
	  "logSize": 2048,
	  "logLevel": "info"
	},
	"blkio":{},
	"qos": {
		"subSys": ["cpu", "memory"]
	},
	"cacheLimit": {
	  "defaultLimitMode": "static",
	  "adjustInterval": 1000,
	  "perfDuration": 1000,
	  "l3Percent": {
		"low": 0,
		"mid": 10,
		"high": 100
	  },
	  "memBandPercent": {
		"low": 10,
		"mid": 30,
		"high": 50
	  }
	}
}
`

func TestServices(t *testing.T) {
	if !util.PathExist(constant.TmpTestDir) {
		if err := os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode); err != nil {
			assert.NoError(t, err)
		}
	}

	defer os.RemoveAll(constant.TmpTestDir)

	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	defer os.Remove(tmpConfigFile)
	if err := ioutil.WriteFile(tmpConfigFile, []byte(rubikConfig), constant.DefaultFileMode); err != nil {
		assert.NoError(t, err)
		return
	}

	c := NewConfig(JSON)
	if err := c.LoadConfig(tmpConfigFile); err != nil {
		assert.NoError(t, err)
		return
	}
	fmt.Printf("config: %v", c)
}
