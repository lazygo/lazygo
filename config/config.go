package config

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"runtime"
)

var conf *config

type config struct {
	file string
	data map[string]gjson.Result
}

func init() {
	conf = &config{
		file: "",
		data: nil,
	}
}

func LoadFile(filename string) error {

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
				return fmt.Errorf("%v: %v", confPath, err)
			}
			continue
		}

		conf.data = gjson.ParseBytes(content).Map()
		conf.file = confPath
		return nil
	}
	return errors.New("未找到配置文件")
}

func GetSection(section string) (*gjson.Result, error) {
	if conf.data == nil {
		return nil, errors.New("未加载配置文件")
	}
	if result, ok := conf.data[section]; ok {
		return &result, nil
	}
	return nil, errors.New("配置段不存在")
}
