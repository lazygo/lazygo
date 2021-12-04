package logger

import (
	"sync"
	"time"
)

type logMsg struct {
	b []byte
	t time.Time
}

var msgPool = &sync.Pool{
	New: func() interface{} {
		return &logMsg{}
	},
}

const defaultAsyncMsgLen = 1e3

type asyncWriter struct {
	lw         logWriter
	msgChanLen int64
	msgChan    chan *logMsg
	signalChan chan string
	wg         sync.WaitGroup
}

// newAsync 异步写入
func newAsync(lw logWriter, chanLens uint64) *asyncWriter {
	a := &asyncWriter{
		lw:         lw,
		signalChan: make(chan string, 1),
	}

	if chanLens <= 0 {
		chanLens = defaultAsyncMsgLen
	}

	a.msgChan = make(chan *logMsg, chanLens)
	a.wg.Add(1)
	go a.start()
	return a
}

func (a *asyncWriter) Write(b []byte, t time.Time) (int, error) {
	msg := msgPool.Get().(*logMsg)
	msg.b = b
	msg.t = t
	a.msgChan <- msg
	return len(b), nil
}

func (a *asyncWriter) Close() {
	a.signalChan <- "close"
	a.wg.Wait()
	close(a.msgChan)
	close(a.signalChan)
}

// start 启动异步
func (a *asyncWriter) start() {
	for {
		if a.lw == nil {
			break
		}
		select {
		case msg := <-a.msgChan:
			_, _ = a.lw.Write(msg.b, msg.t)
			msgPool.Put(msg)
		case sg := <-a.signalChan:
			a.flush()
			if sg == "close" {
				a.lw.Close()
				a.lw = nil
			}
			a.wg.Done()
		}
	}
}

// flush 将channel中数据接收完，并同步logWriter
func (a *asyncWriter) flush() {
	for {
		if len(a.msgChan) > 0 {
			msg := <-a.msgChan
			a.lw.Write(msg.b, msg.t)
			msgPool.Put(msg)
			continue
		}
		break
	}
	a.lw.Flush()
}
