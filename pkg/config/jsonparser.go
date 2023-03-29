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
// Description: This file contains parsing functions for the json language

// Package config is used to manage the configuration of rubik
package config

import (
	"encoding/json"
)

// defaultJsonParser is globally unique json parser
var defaultJsonParser *jsonParser

// jsonParser is used to parse json
type jsonParser struct{}

// getJsonParser gets the globally unique json parser
func getJsonParser() *jsonParser {
	if defaultJsonParser == nil {
		defaultJsonParser = &jsonParser{}
	}
	return defaultJsonParser
}

// ParseConfig parses json data as map[string]interface{}
func (parser *jsonParser) ParseConfig(data []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// UnmarshalSubConfig deserializes interface to structure
func (p *jsonParser) UnmarshalSubConfig(data interface{}, v interface{}) error {
	jsonString, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 1. convert json string to struct
	return json.Unmarshal(jsonString, v)
}

// MarshalIndent serializes interface to string
func (p *jsonParser) MarshalIndent(data interface{}, prefix, indent string) (string, error) {
	if data == nil {
		return "", nil
	}
	res, err := json.MarshalIndent(data, prefix, indent)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
