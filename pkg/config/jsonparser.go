package config

import (
	"encoding/json"
	"fmt"
)

// defaultJsonParser is globally unique json parser
var defaultJsonParser *jsonParser

// jsonParser is used to parse json
type jsonParser struct{}

//getJsonParser gets the globally unique json parser
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
	// 1. convert map[string]interface to json string
	val, ok := data.(map[string]interface{})
	if !ok {
		fmt.Printf("invalid type %T\n", data)
		return fmt.Errorf("invalid type %T", data)
	}
	jsonString, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// 2. convert json string to struct
	return json.Unmarshal(jsonString, v)
}
