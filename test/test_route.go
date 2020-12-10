package main

import (
	"github.com/lazygo/lazygo"
	"github.com/lazygo/lazygo/config"
	"github.com/lazygo/lazygo/test/app"
	"github.com/lazygo/lazygo/utils"
)

func main() {
	r := lazygo.NewRouter()
	r.RegisterController(&app.TestController{})
	config.LoadFile("test")
	conf, err := config.GetSection("server")
	utils.CheckError(err)
	s ,err := lazygo.NewServer(conf, r)
	utils.CheckError(err)
	s.Listen()
}