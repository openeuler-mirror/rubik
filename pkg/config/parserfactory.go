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
// Description: This file contains factory classes for configuring parsers for different languages

// Package config is used to manage the configuration of rubik
package config

type (
	// parserType represents the parser type
	parserType int8
	// parserFactory is the factory class of the parser
	parserFactory struct{}
	// ConfigParser is a configuration parser for different languages
	ConfigParser interface {
		ParseConfig(data []byte) (map[string]interface{}, error)
		UnmarshalSubConfig(data interface{}, v interface{}) error
	}
)

const (
	// JSON represents the json type parser
	JSON parserType = iota
)

// defaultParserFactory is globally unique parser factory
var defaultParserFactory = &parserFactory{}

// getParser gets parser instance according to the parser type passed in
func (factory *parserFactory) getParser(pType parserType) ConfigParser {
	switch pType {
	case JSON:
		return getJsonParser()
	default:
		return getJsonParser()
	}
}
