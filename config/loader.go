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
	data map[string]toml.Primitive
}

func (l *tomlLoader) decode(field string, v interface{}) error {
	data := l.data[field]
	return toml.PrimitiveDecode(data, v)
}

func loadJson(data []byte) (Loader, error) {
	loader := &jsonLoader{}
	err := json.Unmarshal(data, &loader.data)
	return loader, err

}

func loadToml(data []byte) (Loader, error) {
	loader := &tomlLoader{}
	_, err := toml.DecodeReader(bytes.NewReader(data), &loader.data)
	return loader, err
}
