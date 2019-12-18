package lazygo

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"runtime"
)

type Config struct {
	file string
	data map[string]gjson.Result
}

func NewConfig(filename string) (*Config, error) {

	confPaths := make([]string, 0, 3)
	confPaths = append(confPaths, "./"+filename+".json")
	if runtime.GOOS != "windows" {
		confPaths = append(confPaths, "/etc/"+filename+".json")
	}
	confPaths = append(confPaths, "./"+filename+".default.json")

	for _, confPath := range confPaths {
		content, err := ioutil.ReadFile(confPath)
		if err != nil {
			if _, isPathErr := err.(*os.PathError); !isPathErr {
				return nil, fmt.Errorf("%v: %v", confPath, err)
			}
			continue
		}

		config := gjson.ParseBytes(content).Map()
		return &Config{confPath, config}, nil
	}
	return nil, errors.New("未找到配置文件")
}

func (c *Config) GetSection(section string) (*gjson.Result, error) {
	if result, ok := c.data[section]; ok {
		return &result, nil
	}
	return nil, errors.New("配置段不存在")
}
