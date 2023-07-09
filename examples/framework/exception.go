package framework

import (
	"net/http"
	"time"

	"github.com/lazygo/lazygo/server"
)

func HTTPErrorHandlerFunc(err error, ctx Context) {
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

	// Issue #1426
	code := he.Code
	errno := he.Errno
	message := he.Message
	if msg, ok := he.Message.(string); ok {
		if ctx.IsDebug() {
			message = server.Map{
				"code":    code,
				"errno":   errno,
				"message": msg,
				"error":   err.Error(),
				"rid":     ctx.GetRequestID(),
				"t":       time.Now().Unix(),
			}
		} else {
			message = server.Map{
				"code":    code,
				"errno":   errno,
				"message": msg,
				"rid":     ctx.GetRequestID(),
				"t":       time.Now().Unix(),
			}
		}
	}

	// Send response
	if !ctx.ResponseWriter().Committed {
		if ctx.Request().Method == http.MethodHead { // Issue #608
			err = ctx.NoContent(he.Code)
		} else {
			err = ctx.JSON(code, message)
		}
		if err != nil {
			panic(err)
		}
	}
}
