package srv

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/brook/common"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/google/uuid"
	"github.com/panjf2000/gnet/v2"
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

	closeEvents []transport.CloseEvent

	protocol utils.Network
}

func (c *GChannel) SendTo(by []byte, addr net.Addr) (int, error) {
	if c.protocol != utils.NetworkUdp {
		return 0, nil
	}
	return c.Conn.SendTo(by, addr)
}

// SetDeadline is a wrapper for gnet.Conn.SetDeadline.
func (c *GChannel) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *GChannel) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *GChannel) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}

func (c *GChannel) GetConn() net.Conn {
	return c.Conn
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
func (c *GChannel) GetReader() io.Reader {
	return c.Conn
}

// GetWriter
//
//	@Description:
//	@receiver receiver
//	@return aio.Writer
func (c *GChannel) GetWriter() io.Writer {
	return c.Conn
}

// AddHandler
//
//	@Description:
//	@receiver receiver
//	@param handler
func (c *GChannel) AddHandler(handler ...GChannelHandler) {
	c.Handlers = append(c.Handlers, handler...)
}

// GetContext
//
// This function returns the context of the GChannel
func (c *GChannel) GetContext() *ConnContext {
	// Return the context of the GChannel
	return c.Context
}

// OnClose CloseEvent This function takes a pointer to a GChannel and a function as parameters. The function does not return anything.
func (c *GChannel) OnClose(event transport.CloseEvent) {
	c.closeEvents = append(c.closeEvents, event)
}

// Reader
//
//	@Description:
//	@receiver receiver
//	@param out
//	@return int
//	@return error
func (c *GChannel) Read(out []byte) (int, error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	//ErrShortBuffer
	n, err := c.Conn.Read(out)
	if errors.Is(err, io.ErrShortBuffer) {
		//try read.
		if len(out) <= 4 {
			return 0, nil
		}
		return 0, err
	}
	return n, err
}

func (c *GChannel) ReadFull(out []byte) (int, error) {
	return io.ReadFull(c.GetReader(), out)
}

// Writer
//
//	@Description:
//	@receiver receiver
//	@param out
//	@return error
func (c *GChannel) Write(out []byte) (int, error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	return c.Conn.Write(out)
}

// Next
//
//	@Description: Next()
//	@receiver receiver
//	@param pos
//	@return net.Conn
func (c *GChannel) Next(pos int) ([]byte, error) {
	return c.Conn.Next(pos)
}

// GetServer
//
//	@Description:
//	@receiver receiver
//	@return *Server
func (c *GChannel) GetServer() *Server {
	return c.Server
}

// Close
//
//	@Description:
//	@receiver receiver
//	@return error
func (c *GChannel) Close() error {
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
	c.Context.IsClosed = true
	for _, handler := range c.Handlers {
		handler.DoClose(c)
	}
	c.cancel()
	for _, event := range c.closeEvents {
		if c != nil {
			event(c)
		}
	}
	clear(c.closeEvents)
	return nil
}

func (c *GChannel) IsClose() bool {
	if c.Context.IsClosed {
		return true
	}
	select {
	case <-c.Done():
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
func (c *GChannel) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//
// LocalAddr
//  @Description:
//  @receiver receiver
//  @return net.Addr
//

func (c *GChannel) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

// GetNetConn
//
//	@Description:
//	@receiver receiver
//	@return net.Conn
func (c *GChannel) GetNetConn() net.Conn {
	return c.Conn
}

//
// isConnection
//  @Description:
//  @receiver receiver
//  @return bool
//

func (c *GChannel) isConnection() bool {
	return !c.Context.IsClosed
}

// GetAttr
//
//	@Description: Get conn attr value.
//	@receiver receiver
//	@param key
//	@return interface{}
//	@return bool
func (c *GChannel) GetAttr(key common.KeyType) (interface{}, bool) {
	i, ok := c.Context.attr[key]
	return i, ok
}

// ConnContext
// @Description: connContext info.
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

func (c *GChannel) GetId() string {
	return c.Id
}

func (c *GChannel) Done() <-chan struct{} {
	return c.bgCtx.Done()
}
