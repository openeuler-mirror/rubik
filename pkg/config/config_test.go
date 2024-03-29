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

func TestNewConfig(t *testing.T) {
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
	},
	"ioCost": [
	  {
		"nodeName": "k8s-single",
		"config": [
		  {
			"dev": "sdb",
			"enable": true,
			"model": "linear",
			"param": {
			  "rbps": 10000000,
			  "rseqiops": 10000000,
			  "rrandiops": 10000000,
			  "wbps": 10000000,
			  "wseqiops": 10000000,
			  "wrandiops": 10000000
			}
		  }
		]
	  }
	],
	"psi": {
		"interval": 10,
		"resource": [
		  "cpu",
		  "memory",
		  "io"
		],
		"avg10Threshold": 5.0
	}
}
`
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
	}

	c := NewConfig(JSON)
	if err := c.LoadConfig(tmpConfigFile); err != nil {
		assert.NoError(t, err)
	}
	fmt.Printf("config: %v", c)
}

func TestNewConfigNoConfig(t *testing.T) {
	c := &Config{}
	if err := c.LoadConfig(""); err == nil {
		t.Fatalf("Config file exists")
	}
}

func TestNewConfigDamagedConfig(t *testing.T) {
	var rubikConfig string = `{`
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
	}

	c := NewConfig(JSON)
	if err := c.LoadConfig(tmpConfigFile); err == nil {
		t.Fatalf("Damaged config file should not be loaded.")
	}
}

func TestNewConfigNoAgentConfig(t *testing.T) {
	var rubikConfig string = `{}`
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
	}

	c := NewConfig(JSON)
	if err := c.LoadConfig(tmpConfigFile); err != nil {
		assert.NoError(t, err)
	}
	fmt.Printf("config: %v", c)
}

func TestUnwrapServiceConfig(t *testing.T) {
	c := &Config{}
	c.Fields = make(map[string]interface{})
	c.Fields["agent"] = nil
	c.Fields["config"] = nil
	sc := c.UnwrapServiceConfig()
	if _, exist := sc["agent"]; exist {
		t.Fatalf("agent is exists")
	}
	if _, exist := sc["config"]; !exist {
		t.Fatalf("config is not exists")
	}
}
