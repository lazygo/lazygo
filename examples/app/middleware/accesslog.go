package middleware

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo/lazygo/server"
	"github.com/lazygo/pkg/monitor"
	"github.com/shirou/gopsutil/v4/process"
)

var rnum int32 = 0

// AccessLog 访问日志记录中间件
func AccessLog(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		uri := ctx.Request().RequestURI

		errno := 0
		code := ctx.ResponseWriter().Status
		st := time.Now()

		if ctx.IsDebug() {
			req := ctx.Request()
			buffer := &bytes.Buffer{}
			reader := io.TeeReader(req.Body, buffer)
			body, err := io.ReadAll(reader)
			if err != nil {
				ctx.Logger().Error(">>> :%+v\n, err:%+v", body, err)
			}
			ctx.Logger().Debug(">>> body:%+v", string(body))

			req.Body = io.NopCloser(buffer)
		}

		atomic.AddInt32(&rnum, 1)
		defer atomic.AddInt32(&rnum, -1)

		defer func() {
			rec := recover()
			if rec != nil {
				ctx.Logger().Alert("%v", rec)
				ctx.Logger().Error("%s", string(debug.Stack()))
				framework.Server().HTTPErrorHandler(fmt.Errorf("%v", rec), ctx)
				errno = 500
			}

			sysInfo, err := monitor.ReportSysMonitor(ctx)
			if err != nil {
				ctx.Logger().Warn("report sys monitor fail %v", err)
				sysInfo = &monitor.SysMontor{MemInfo: &process.MemoryInfoStat{}}
			}
			ctx.Logger().Notice(
				"[pid: %d] [threads: %d] [goroutine: %d] [mem: %.2fM %.1f%%] [cpu: %.1f%%] [fds: %d] [rnum: %d] [time: %.1fms] [status: %d] [errno: %d] [ip: %s] [request_uri: %s]",
				os.Getegid(),
				sysInfo.NumThreads,
				sysInfo.NumGoroutine,
				float64(sysInfo.MemInfo.RSS)/1024/1024,
				sysInfo.MemPercent,
				sysInfo.CPUPercent,
				sysInfo.NumFDs,
				atomic.LoadInt32(&rnum),
				float64(time.Since(st).Microseconds())/1000,
				code,
				errno,
				ctx.RealIP(),
				uri,
			)
		}()

		respErr := next(ctx)
		if respErr != nil {
			if he, ok := respErr.(*server.HTTPError); ok {
				errno = he.Errno
				code = he.Code
			}
		}

		return respErr
	})
}
