package cos

import (
	"errors"

	"github.com/lazygo/lazygo/libhttp"
)

var client *libhttp.HttpClient

func Init(conf Config) error {
	client = libhttp.NewClient(conf.BucketURL, "POST").
		AddParam("key", conf.SecretID).
		AddParam("secret", conf.SecretKey)
	return nil
}

func Upload(name string, filename string) error {
	if client == nil {
		return errors.New("cos 未初始化")
	}
	var resp struct {
		Errno int    `json:"errno"`
		Msg   string `json:"msg"`
	}
	err := client.PostFile(name, filename).ToJSON(&resp)
	if err != nil {
		return err
	}
	if resp.Errno != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}
