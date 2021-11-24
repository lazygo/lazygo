package logger

import (
	"sync"
	"time"
)

type msg struct {
	b []byte
	t time.Time
}

var msgPool = &sync.Pool{
	New: func() interface{} {
		return &msg{}
	},
}

const defaultAsyncMsgLen = 1e3

type asyncWriter struct {
	lw         logWriter
	msgChanLen int64
	msgChan    chan *msg
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

	a.msgChan = make(chan *msg, chanLens)
	a.wg.Add(1)
	go a.start()
	return a
}

func (a *asyncWriter) Writeln(b []byte, t time.Time) (int, error) {
	return a.lw.Writeln(b, t)
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
			_, _ = a.lw.Writeln(msg.b, msg.t)
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
			a.lw.Writeln(msg.b, msg.t)
			msgPool.Put(msg)
			continue
		}
		break
	}
	a.lw.Flush()
}
