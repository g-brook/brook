package srv

import (
	"time"

	"github.com/brook/common"
	"github.com/google/uuid"
)

type ConnContext struct {
	IsClosed   bool
	Id         string
	lastActive time.Time
	IsTimeOut  bool
	attr       map[common.KeyType]interface{}
	isSmux     bool
}

func NewConnContext(isUdp bool, addr string) *ConnContext {
	var id string
	if isUdp {
		id = addr
	} else {
		id = uuid.New().String()
	}
	return &ConnContext{
		IsClosed:   false,
		Id:         id,
		lastActive: time.Now(),
		IsTimeOut:  false,
		attr:       make(map[common.KeyType]interface{}),
		isSmux:     false,
	}
}

type GContext interface {
	GetContext() *ConnContext
	Next(pos int) ([]byte, error)
}
