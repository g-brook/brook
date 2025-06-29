package srv

import (
	"context"
	"errors"
	"github.com/RussellLuo/timingwheel"
	"github.com/brook/common"
	"github.com/brook/common/transport"
	"github.com/google/uuid"
	"github.com/panjf2000/gnet/v2"
	"io"
	"net"
	"time"
)

// GChannel
// @Description:
type GChannel struct {
	Conn gnet.Conn

	Id string

	Context *ConnContext

	Server *Server

	Handlers []GChannelHandler

	PipeConn *transport.SmuxAdapterConn

	bgCtx context.Context

	cancel context.CancelFunc
}

func (receiver *GChannel) SetDeadline(t time.Time) error {
	return receiver.Conn.SetDeadline(t)
}

func (receiver *GChannel) SetReadDeadline(t time.Time) error {
	return receiver.Conn.SetReadDeadline(t)
}

func (receiver *GChannel) SetWriteDeadline(t time.Time) error {
	return receiver.Conn.SetWriteDeadline(t)
}

func (receiver *GChannel) GetConn() net.Conn {
	return receiver.Conn
}

// GChannelHandler
// @Description:
type GChannelHandler interface {
	DoOpen(conn *GChannel)

	DoClose(conn *GChannel)
}

// GetReader
//
//	@Description: Get from gnet.conn.
//	@receiver receiver
//	@return aio.Reader
func (receiver *GChannel) GetReader() io.Reader {
	return receiver.Conn
}

// GetWriter
//
//	@Description:
//	@receiver receiver
//	@return aio.Writer
func (receiver *GChannel) GetWriter() io.Writer {
	return receiver.Conn
}

// AddHandler
//
//	@Description:
//	@receiver receiver
//	@param handler
func (receiver *GChannel) AddHandler(handler ...GChannelHandler) {
	receiver.Handlers = append(receiver.Handlers, handler...)
}

// GetContext
//
//	@Description:
//	@receiver receiver
//	@return *ConnContext
func (receiver *GChannel) GetContext() *ConnContext {
	return receiver.Context
}

// Reader
//
//	@Description:
//	@receiver receiver
//	@param out
//	@return int
//	@return error
func (receiver *GChannel) Read(out []byte) (int, error) {
	//ErrShortBuffer
	n, err := receiver.Conn.Read(out)
	if errors.Is(err, io.ErrShortBuffer) {
		return 0, nil
	}
	return n, err
}

func (receiver *GChannel) ReadFull(out []byte) (int, error) {
	return io.ReadFull(receiver.GetReader(), out)
}

// Writer
//
//	@Description:
//	@receiver receiver
//	@param out
//	@return error
func (receiver *GChannel) Write(out []byte) (int, error) {
	return receiver.Conn.Write(out)
}

// Next
//
//	@Description: Next()
//	@receiver reveiver
//	@param pos
//	@return net.Conn
func (receiver *GChannel) Next(pos int) ([]byte, error) {
	return receiver.Conn.Next(pos)
}

// GetServer
//
//	@Description:
//	@receiver receiver
//	@return *Server
func (receiver *GChannel) GetServer() *Server {
	return receiver.Server
}

// Close
//
//	@Description:
//	@receiver receiver
//	@return error
func (receiver *GChannel) Close() error {
	if receiver.Context.Timer != nil {
		receiver.Context.Timer.Stop()
	}
	if receiver.Conn != nil {
		_ = receiver.Conn.Close()
	}
	receiver.Context.IsClosed = true
	for _, handler := range receiver.Handlers {
		handler.DoClose(receiver)
	}
	receiver.cancel()
	return nil
}

func (s *GChannel) IsClose() bool {
	select {
	case <-s.Done():
		return true
	default:
		return false
	}
}

// RemoteAddr
//
//	@Description:
//	@receiver receiver
//	@return net.Addr
func (receiver *GChannel) RemoteAddr() net.Addr {
	return receiver.Conn.RemoteAddr()
}

//
// LocalAddr
//  @Description:
//  @receiver receiver
//  @return net.Addr
//

func (receiver *GChannel) LocalAddr() net.Addr {
	return receiver.Conn.LocalAddr()
}

// GetNetConn
//
//	@Description:
//	@receiver receiver
//	@return net.Conn
func (receiver *GChannel) GetNetConn() net.Conn {
	return receiver.Conn
}

//
// isConnection
//  @Description:
//  @receiver receiver
//  @return bool
//

func (receiver *GChannel) isConnection() bool {
	return !receiver.Context.IsClosed
}

// GetAttr
//
//	@Description: Get conn attr value.
//	@receiver receiver
//	@param key
//	@return interface{}
//	@return bool
func (receiver *GChannel) GetAttr(key common.KeyType) (interface{}, bool) {
	i, ok := receiver.Context.attr[key]
	return i, ok
}

// ConnContext
// @Description: connContext info.
type ConnContext struct {
	IsClosed   bool
	Id         string
	lastActive time.Time
	IsTimeOut  bool
	Timer      *timingwheel.Timer
	attr       map[common.KeyType]interface{}
	isSmux     bool
}

func NewConnContext() *ConnContext {
	return &ConnContext{
		IsClosed:   false,
		Id:         uuid.New().String(),
		lastActive: time.Now(),
		IsTimeOut:  false,
		attr:       make(map[common.KeyType]interface{}),
		isSmux:     false,
	}
}

// AddAttr
//
//	@Description: Add a attr info on Conn.
//	@receiver receiver
func (receiver *ConnContext) AddAttr(key common.KeyType, value interface{}) {
	receiver.attr[key] = value
}

// GetAttr
//
//	@Description: Get conn attr value.
//	@receiver receiver
//	@param key
//	@return interface{}
//	@return bool
func (receiver *ConnContext) GetAttr(key common.KeyType) (interface{}, bool) {
	i, ok := receiver.attr[key]
	return i, ok
}

// LastActive
//
//	@Description:
//	@receiver receiver
func (receiver *ConnContext) LastActive() {
	receiver.lastActive = time.Now()
	//conn.context = receiver
}

// GetLastActive
//
//	@Description:
//	@receiver receiver
//	@return time.Time
func (receiver *ConnContext) GetLastActive() time.Time {
	return receiver.lastActive
}

func (receiver *GChannel) GetId() string {
	return receiver.Id
}

func (receiver *GChannel) Done() <-chan struct{} {
	return receiver.bgCtx.Done()
}
