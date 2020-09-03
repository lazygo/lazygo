package lazygo

import (
	"fmt"
	"github.com/lazygo/lazygo/utils"
	"github.com/tidwall/gjson"
	"net/http"
	"strconv"
)

type Server struct {
	host   string
	port   int
	router *Router
}

/*
{"host": "127.0.0.1", "port": 8080}
*/

func NewServer(conf *gjson.Result, router *Router) (*Server, error) {
	host := conf.Get("host").String()
	port := conf.Get("port").Int()

	return &Server{
		host:   utils.ToString(host, "127.0.0.1"),
		port:   utils.ToInt(port, 8080),
		router: router,
	}, nil
}

func (s *Server) Listen() {
	fmt.Print(s.host + ":" + strconv.Itoa(s.port))
	http.ListenAndServe(s.host+":"+strconv.Itoa(s.port), s.router.GetHandle())
}
