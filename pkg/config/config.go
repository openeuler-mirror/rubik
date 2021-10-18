// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Haomin Tsai
// Create: 2021-09-28
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
	AutoCheck  bool   `json:"autoCheck,omitempty"`
	LogDriver  string `json:"logDriver,omitempty"`
	LogDir     string `json:"logDir,omitempty"`
	LogSize    int    `json:"logSize,omitempty"`
	LogLevel   string `json:"logLevel,omitempty"`
	CgroupRoot string `json:"cgroupRoot,omitempty"`
}

// NewConfig returns new config load from config file
func NewConfig(path string) (*Config, error) {
	if path == "" {
		path = constant.ConfigFile
	}

	defaultLogSize := 1024
	cfg := Config{
		AutoCheck:  false,
		LogDriver:  "stdio",
		LogDir:     constant.DefaultLogDir,
		LogSize:    defaultLogSize,
		LogLevel:   "info",
		CgroupRoot: constant.DefaultCgroupRoot,
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
