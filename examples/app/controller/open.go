package controller

import (
	request "github.com/lazygo-dev/lazygo/examples/app/request/open"
	"github.com/lazygo-dev/lazygo/examples/framework"
)

type OpenColtroller struct {
	Ctx framework.Context
}

func (ctl *OpenColtroller) Index(req *request.IndexRequest) error {
	return nil
}
