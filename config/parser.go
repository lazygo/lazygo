package config

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
)

type Parser interface {
	decode(field string, v interface{}) error
}

type jsonParser struct {
	data map[string]json.RawMessage
}

func (l *jsonParser) decode(field string, v interface{}) error {
	data := l.data[field]
	return json.Unmarshal(data, v)
}

type tomlParser struct {
	metadata toml.MetaData
	data     map[string]toml.Primitive
}

func (l *tomlParser) decode(field string, v interface{}) error {
	data := l.data[field]
	return l.metadata.PrimitiveDecode(data, v)
}

func loadJson(data []byte) (Parser, error) {
	parser := &jsonParser{}
	err := json.Unmarshal(data, &parser.data)
	return parser, err

}

func loadToml(data []byte) (Parser, error) {
	parser := &tomlParser{}
	metadata, err := toml.DecodeReader(bytes.NewReader(data), &parser.data)
	parser.metadata = metadata
	return parser, err
}
