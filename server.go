package lazygo

import (
	"fmt"
	"github.com/lazygo/lazygo/library"
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
		host:   library.ToString(host, "127.0.0.1"),
		port:   library.ToInt(port, 8080),
		router: router,
	}, nil
}

func (s *Server) Listen() {
	fmt.Print(s.host + ":" + strconv.Itoa(s.port))
	http.ListenAndServe(s.host+":"+strconv.Itoa(s.port), s.router.GetHandle())
}
