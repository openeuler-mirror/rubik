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
// Description: This file contains configuration content and provides external interaction functions

// Package config is used to manage the configuration of rubik
package config

import (
	"context"
	"fmt"
	"io/ioutil"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/services"
)

const agentKey = "agent"

// sysConfKeys saves the system configuration key, which is the service name except
var sysConfKeys = map[string]struct{}{
	agentKey: {},
}

// Config saves all configuration information of rubik
type Config struct {
	api.ConfigParser
	Agent  *AgentConfig
	Fields map[string]interface{}
}

// AgentConfig is the configuration of rubik, including important basic configurations such as logs
type AgentConfig struct {
	LogDriver  string `json:"logDriver,omitempty"`
	LogLevel   string `json:"logLevel,omitempty"`
	LogSize    int64  `json:"logSize,omitempty"`
	LogDir     string `json:"logDir,omitempty"`
	CgroupRoot string `json:"cgroupRoot,omitempty"`
}

// NewConfig returns an config object pointer
func NewConfig(pType parserType) *Config {
	c := &Config{
		ConfigParser: defaultParserFactory.getParser(pType),
		Agent: &AgentConfig{
			LogDriver:  constant.DefaultLogDir,
			LogSize:    constant.DefaultLogSize,
			LogLevel:   constant.DefaultLogLevel,
			LogDir:     constant.DefaultLogDir,
			CgroupRoot: constant.DefaultCgroupRoot,
		},
	}
	return c
}

// loadConfigFile loads data from configuration file
func loadConfigFile(config string) ([]byte, error) {
	buffer, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// parseAgentConfig parses config as AgentConfig
func (c *Config) parseAgentConfig() {
	content, ok := c.Fields[agentKey]
	if !ok {
		return
	}
	c.UnmarshalSubConfig(content, c.Agent)
}

// LoadConfig loads and parses configuration data from the file, and save it to the Config
func (c *Config) LoadConfig(path string) error {
	if path == "" {
		path = constant.ConfigFile
	}
	data, err := loadConfigFile(path)
	if err != nil {
		return fmt.Errorf("error loading config file %s: %w", path, err)
	}
	fields, err := c.ParseConfig(data)
	if err != nil {
		return fmt.Errorf("error parsing data: %s", err)
	}
	c.Fields = fields
	c.parseAgentConfig()
	if err := c.PrepareServices(); err != nil {
		return fmt.Errorf("error preparing services: %s", err)
	}
	return nil
}

// filterNonServiceKeys returns true when inputting a non-service name
func (c *Config) filterNonServiceKeys(name string) bool {
	// 1. ignore system configured key values
	_, ok := sysConfKeys[name]
	return ok
}

// PrepareServices parses the to-be-run services config and loads them to the ServiceManager
func (c *Config) PrepareServices() error {
	// TODO: Later, consider placing the function book in rubik.go
	for name, config := range c.Fields {
		if c.filterNonServiceKeys(name) {
			continue
		}
		creator := services.GetServiceCreator(name)
		if creator == nil {
			return fmt.Errorf("no corresponding module %v, please check the module name", name)
		}
		service := creator()
		if err := c.UnmarshalSubConfig(config, service); err != nil {
			return fmt.Errorf("error unmarshaling %s config: %v", name, err)
		}

		if services.SetLoggerOnService(service,
			log.WithCtx(context.WithValue(context.Background(), log.CtxKey(constant.LogEntryKey), name))) {
			log.Debugf("set logger for service: %s", name)
		}

		// try to verify configuration
		if validator, ok := service.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("error configuring service %s: %v", name, err)
			}
		}
		services.AddRunningService(name, service)
	}
	return nil
}

// Validator is a function interface to verify whether the configuration is correct or not
type Validator interface {
	Validate() error
}
