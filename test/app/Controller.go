package app

import (
	"errors"
	"github.com/lazygo/lazygo"
)

type Controller struct {
	lazygo.Controller
}

var err = errors.New("aaaadadaaa")
var err2 = errors.New("34532412534")

func (ctl *Controller) Init()  {
	ctl.ErrorHandler(err, func(err error) {
		ctl.ApiFail(1, err.Error(), nil)
	})
	ctl.ErrorHandler(err2, func(err error) {
		ctl.ApiFail(1, err.Error(), nil)
	})
	ctl.InitTpl("templates/", ".html", nil)
}