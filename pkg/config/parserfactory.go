package config

import "isula.org/rubik/pkg/api"

type (
	// parserType represents the parser type
	parserType int8
	// parserFactory is the factory class of the parser
	parserFactory struct{}
)

const (
	// JSON represents the json type parser
	JSON parserType = iota
)

// defaultParserFactory is globally unique parser factory
var defaultParserFactory = &parserFactory{}

// getParser gets parser instance according to the parser type passed in
func (factory *parserFactory) getParser(pType parserType) api.ConfigParser {
	switch pType {
	case JSON:
		return getJsonParser()
	default:
		return getJsonParser()
	}
}
