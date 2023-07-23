package config

import (
	"errors"
	"reflect"
)

type Loader func(data []byte) (*Config, error)

func Toml(data []byte) (*Config, error) {
	parser, err := loadToml(data)
	if err != nil {
		return nil, err
	}
	config := &Config{
		parser: parser,
	}
	return config, err
}

func Json(data []byte) (*Config, error) {
	parser, err := loadJson(data)
	if err != nil {
		return nil, err
	}

	config := &Config{
		parser: parser,
	}
	return config, err
}

type Config struct {
	parser Parser
}

// Load 加载一段配置
func (c *Config) Load(field string, f interface{}) error {
	rf := reflect.ValueOf(f)
	tf := rf.Type()
	if tf.NumIn() != 1 {
		return errors.New("")
	}
	if tf.NumOut() != 1 {
		return errors.New("")
	}
	pt := tf.In(0)

	pv := reflect.New(pt)
	err := c.parser.decode(field, pv.Interface())
	if err != nil {
		return err
	}

	out := rf.Call([]reflect.Value{pv.Elem()})

	if ierr := out[0].Interface(); ierr != nil {
		if err = ierr.(error); err != nil {
			return err
		}
	}
	return nil
}
