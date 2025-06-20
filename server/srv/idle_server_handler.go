package srv

import (
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
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

func (b *IdleServerHandler) Open(conn *GChannel, traverse TraverseBy) {
	if b.timeout > 0 {
		timer := utils.NewWheel.ScheduleFunc(&TimeoutScheduler{
			timeout: b.timeout,
		}, func() {
			conn.GetContext().IsTimeOut = time.Now().After(conn.GetContext().GetLastActive())
			if conn.GetContext().IsTimeOut {
				log.Warn("Connection timeout:  %s -> %s", conn.LocalAddr().String(), conn.RemoteAddr().String())
				_ = conn.Conn.Wake(nil)
			}
		})
		conn.Context.Timer = timer
	}
	traverse()
}

func (b *IdleServerHandler) Reader(conn *GChannel, traverse TraverseBy) {
	if conn.GetContext().IsTimeOut {
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
