// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-05-07
// Description: config load test

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/constant"
)

var (
	cfgA = `
{
	"autoCheck": true,
	"logDriver": "file",
	"logDir": "/tmp/rubik-test",
	"logSize": 2048,
	"logLevel": "debug",
	"cgroupRoot": "/tmp/rubik-test/cgroup"
}`
)

// TestNewConfig is NewConfig function test
func TestNewConfig(t *testing.T) {
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	os.Remove(tmpConfigFile)

	// coverage
	NewConfig("")

	// test_rubik_load_config_file_0001
	defaultLogSize := 1024
	cfg, err := NewConfig(tmpConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, cfg.AutoCheck, false)
	assert.Equal(t, cfg.LogDriver, "stdio")
	assert.Equal(t, cfg.LogDir, constant.DefaultLogDir)
	assert.Equal(t, cfg.LogSize, defaultLogSize)
	assert.Equal(t, cfg.LogLevel, "info")
	assert.Equal(t, cfg.CgroupRoot, constant.DefaultCgroupRoot)

	// test_rubik_load_config_file_0003
	err = os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)

	logSize := 2048
	err = ioutil.WriteFile(tmpConfigFile, []byte(cfgA), constant.DefaultFileMode)
	assert.NoError(t, err)
	cfg, err = NewConfig(tmpConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, cfg.AutoCheck, true)
	assert.Equal(t, cfg.LogDriver, "file")
	assert.Equal(t, cfg.LogDir, "/tmp/rubik-test")
	assert.Equal(t, cfg.LogSize, logSize)
	assert.Equal(t, cfg.LogLevel, "debug")
	assert.Equal(t, cfg.CgroupRoot, "/tmp/rubik-test/cgroup")

	// test_rubik_load_config_file_0002
	err = ioutil.WriteFile(tmpConfigFile, []byte("abc"), constant.DefaultFileMode)
	assert.NoError(t, err)
	_, err = NewConfig(tmpConfigFile)
	assert.Contains(t, err.Error(), "invalid character")

	size := 20000000
	big := make([]byte, size, size)
	err = ioutil.WriteFile(tmpConfigFile, big, constant.DefaultFileMode)
	assert.NoError(t, err)
	_, err = NewConfig(tmpConfigFile)
	assert.Contains(t, err.Error(), "too big")

	err = os.Remove(tmpConfigFile)
	assert.NoError(t, err)
}

// TestConfig_String is config string function test
func TestConfig_String(t *testing.T) {
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	os.Remove(tmpConfigFile)

	cfg, err := NewConfig(tmpConfigFile)
	assert.NoError(t, err)

	cfgString := cfg.String()
	assert.Equal(t, cfgString, `{
    "logDriver": "stdio",
    "logDir": "/var/log/rubik",
    "logSize": 1024,
    "logLevel": "info",
    "cgroupRoot": "/sys/fs/cgroup",
    "cacheConfig": {
        "defaultLimitMode": "static",
        "adjustInterval": 1000,
        "perfDuration": 1000,
        "l3Percent": {
            "low": 20,
            "mid": 30,
            "high": 50
        },
        "memBandPercent": {
            "low": 10,
            "mid": 30,
            "high": 50
        }
    },
    "blkConfig": {},
    "memConfig": {
        "strategy": "none",
        "checkInterval": 5
    }
}`)
}
