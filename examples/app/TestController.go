package app

import (
	"fmt"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/server"
)

type TestController server.Controller

func (t TestController) TestResponseAction(ctx server.Context) error {
	//fmt.Println(ctx.Param("xxx"))
	db, err := mysql.Database("test")
	if err != nil {
		fmt.Println(err)
	}
	var s []int
	for i := 0; i < 100000; i++ {
		s = append(s, i)
	}
	cond := map[string]interface{}{
		"id": 2,
	}
	data, err := db.Table("test").Where(cond).FetchRow([]interface{}{"*"})
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(data["name"])
	return ctx.JSON(200, map[string]interface{}{
		"aa": data,
		"l":  len(s),
	})
}
