package srv

import (
	"github.com/brook/common/log"
	"time"
)

type IdleServerHandler struct {
	BaseServerHandler

	timeout time.Duration
}

func NewIdleServerHandler(readerTime time.Duration) *IdleServerHandler {
	return &IdleServerHandler{
		timeout: readerTime,
	}
}

func (b *IdleServerHandler) Open(conn *ConnV2, traverse TraverseBy) {
	if b.timeout > 0 {
		timer := newWheel.ScheduleFunc(&TimeoutScheduler{
			timeout: b.timeout,
		}, func() {
			conn.GetContext().isTimeOut = time.Now().After(conn.GetContext().GetLastActive())
			if conn.GetContext().isTimeOut {
				log.Warn("Connection timeout:  %s -> %s", conn.LocalAddr().String(), conn.RemoteAddr().String())
				_ = conn.conn.Wake(nil)
			}
		})
		conn.context.timer = timer
	}
	traverse()
}

func (b *IdleServerHandler) Reader(conn *ConnV2, traverse TraverseBy) {
	if conn.GetContext().isTimeOut {
		log.Warn("Timeout close connection:%s -> %s", conn.LocalAddr().String(), conn.RemoteAddr().String())
		_ = conn.Close()
		return
	}
	traverse()
}

type TimeoutScheduler struct {
	timeout time.Duration
}

func (t *TimeoutScheduler) Next(t2 time.Time) time.Time {
	return t2.Add(t.timeout)
}
