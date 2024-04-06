package framework

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lazygo/lazygo/server"
)

func AppHTTPErrorHandler(err error, ctx Context) {
	ctx.Logger().Error("http return error: url: %s, headers: %+v, err: %+v", ctx.Request().RequestURI, ctx.Request().Header, err)
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

	code := he.Code
	errno := he.Errno
	message := he.Message

	switch msg := he.Message.(type) {
	case string:
		if ctx.IsDebug() {
			message = server.Map{
				"code":  code,
				"errno": errno,
				"msg":   msg,
				"error": err.Error(),
				"rid":   ctx.RequestID(),
				"t":     time.Now().Unix(),
			}
		} else {
			message = server.Map{
				"code":  code,
				"errno": errno,
				"msg":   msg,
				"rid":   ctx.RequestID(),
				"t":     time.Now().Unix(),
			}
		}
	case json.Marshaler:
		// do nothing - this type knows how to format itself to JSON
	case error:
		message = server.Map{
			"code":  code,
			"errno": errno,
			"msg":   msg.Error(),
			"rid":   ctx.RequestID(),
			"t":     time.Now().Unix(),
		}
	}

	// Send response
	if ctx.Request().Method == http.MethodHead {
		err = ctx.NoContent(he.Code)
	} else {
		err = ctx.JSON(code, message)
	}
	if err != nil {
		ctx.Logger().Error("", err)
	}

}
