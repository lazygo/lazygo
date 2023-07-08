package config

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
)

type Loader interface {
	decode(field string, v interface{}) error
}

type jsonLoader struct {
	data map[string]json.RawMessage
}

func (l *jsonLoader) decode(field string, v interface{}) error {
	data := l.data[field]
	return json.Unmarshal(data, v)
}

type tomlLoader struct {
	metadata toml.MetaData
	data     map[string]toml.Primitive
}

func (l *tomlLoader) decode(field string, v interface{}) error {
	data := l.data[field]
	return l.metadata.PrimitiveDecode(data, v)
}

func loadJson(data []byte) (Loader, error) {
	loader := &jsonLoader{}
	err := json.Unmarshal(data, &loader.data)
	return loader, err

}

func loadToml(data []byte) (Loader, error) {
	loader := &tomlLoader{}
	metadata, err := toml.DecodeReader(bytes.NewReader(data), &loader.data)
	loader.metadata = metadata
	return loader, err
}
