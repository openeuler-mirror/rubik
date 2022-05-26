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
// Create: 2021-04-26
// Description: config load

package config

import (
	"bytes"
	"encoding/json"
	"path/filepath"

	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/util"
)

var (
	// CgroupRoot is cgroup mount point
	CgroupRoot = constant.DefaultCgroupRoot
	// ShutdownFlag is rubik shutdown flag
	ShutdownFlag int32
	// ShutdownChan is rubik shutdown channel
	ShutdownChan = make(chan struct{})
)

// Config defines the configuration for rubik
type Config struct {
	AutoConfig bool        `json:"autoConfig,omitempty"`
	AutoCheck  bool        `json:"autoCheck,omitempty"`
	LogDriver  string      `json:"logDriver,omitempty"`
	LogDir     string      `json:"logDir,omitempty"`
	LogSize    int         `json:"logSize,omitempty"`
	LogLevel   string      `json:"logLevel,omitempty"`
	CgroupRoot string      `json:"cgroupRoot,omitempty"`
	CacheCfg   CacheConfig `json:"cacheConfig,omitempty"`
	BlkConfig  BlkioConfig `json:"blkConfig,omitempty"`
}

// CacheConfig define cache limit related config
type CacheConfig struct {
	Enable            bool            `json:"enable,omitempty"`
	DefaultLimitMode  string          `json:"defaultLimitMode,omitempty"`
	DefaultResctrlDir string          `json:"-"`
	AdjustInterval    int             `json:"adjustInterval,omitempty"`
	PerfDuration      int             `json:"perfDuration,omitempty"`
	L3Percent         MultiLvlPercent `json:"l3Percent,omitempty"`
	MemBandPercent    MultiLvlPercent `json:"memBandPercent,omitempty"`
}

// BlkioConfig defines blkio related configurations.
type BlkioConfig struct {
	Limit bool `json:"limit,omitempty"`
}

// MultiLvlPercent define multi level percentage
type MultiLvlPercent struct {
	Low  int `json:"low,omitempty"`
	Mid  int `json:"mid,omitempty"`
	High int `json:"high,omitempty"`
}

// NewConfig returns new config load from config file
func NewConfig(path string) (*Config, error) {
	if path == "" {
		path = constant.ConfigFile
	}

	defaultLogSize, defaultAdInt, defaultPerfDur := 1024, 1000, 1000
	defaultLowL3, defaultMidL3, defaultHighL3, defaultLowMB, defaultMidMB, defaultHighMB := 20, 30, 50, 10, 30, 50
	cfg := Config{
		AutoCheck:  false,
		LogDriver:  "stdio",
		LogDir:     constant.DefaultLogDir,
		LogSize:    defaultLogSize,
		LogLevel:   "info",
		CgroupRoot: constant.DefaultCgroupRoot,
		CacheCfg: CacheConfig{
			Enable:            false,
			DefaultLimitMode:  "static",
			DefaultResctrlDir: "/sys/fs/resctrl",
			AdjustInterval:    defaultAdInt,
			PerfDuration:      defaultPerfDur,
			L3Percent: MultiLvlPercent{
				Low:  defaultLowL3,
				Mid:  defaultMidL3,
				High: defaultHighL3,
			},
			MemBandPercent: MultiLvlPercent{
				Low:  defaultLowMB,
				Mid:  defaultMidMB,
				High: defaultHighMB,
			},
		},
		BlkConfig: BlkioConfig{
			Limit: false,
		},
	}

	defer func() {
		CgroupRoot = cfg.CgroupRoot
	}()

	if !util.PathExist(path) {
		return &cfg, nil
	}

	b, err := util.ReadSmallFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(b)
	if err := json.NewDecoder(reader).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// String return string format.
func (cfg *Config) String() string {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
