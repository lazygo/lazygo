package framework

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lazygo/lazygo/server"
)

func AppHTTPErrorHandler(err error, ctx Context) {
	ctx.Logger().Notice("[http return error] [url: %s] [headers: %+v] [err: %+v]", ctx.Request().RequestURI, ctx.Request().Header, err)
	if ctx.ResponseWriter().Committed {
		return
	}
	he, ok := err.(*server.HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*server.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &server.HTTPError{
			Code:    http.StatusInternalServerError,
			Errno:   http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	resp := Response[any]{
		Code:  he.Code,
		Errno: he.Errno,
		Rid:   ctx.RequestID(),
		Time:  time.Now().Unix(),
		Data:  nil,
	}

	switch msg := he.Message.(type) {
	case string:
		resp.Msg = msg
		if ctx.IsDebug() {
			resp.Error = err.Error()
		}
	case json.Marshaler:
		// do nothing - this type knows how to format itself to JSON
	case error:
		resp.Error = msg.Error()
	}

	// Send response
	if ctx.Request().Method == http.MethodHead {
		err = ctx.NoContent(he.Code)
	} else {
		err = ctx.JSON(200, resp)
	}
	if err != nil {
		ctx.Logger().Error("%v", err)
	}

}
