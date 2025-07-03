package clis

import (
	"github.com/brook/common/log"
	"sync"
	"time"
)

type ReconnectFunction func() bool
type ReconnectManager struct {
	timer             *time.Timer
	reconnectInterval time.Duration
	retries           int
	isStart           bool
	lock              sync.Mutex
}

func NewReconnectionManager(t time.Duration) *ReconnectManager {
	return &ReconnectManager{
		timer:             time.NewTimer(t),
		reconnectInterval: t,
	}
}

func (r *ReconnectManager) tryReconnect(rf ReconnectFunction) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.isStart {
		return
	}
	r.isStart = true
	r.timer.Reset(r.reconnectInterval)
	go func() {
		for {
			select {
			case <-r.timer.C:
				r.retries++
				log.Info("Try reconnect %v count, now.", r.retries)
				b := rf()
				if b {
					r.isStart = false
					r.retries = 0
					return
				}
				r.timer.Reset(r.reconnectInterval)
			}

		}
	}()
}
