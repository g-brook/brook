package remote

import (
	"context"
	"fmt"
	"github.com/brook/common/log"
	"github.com/xtaci/smux"
	"net"
	"sync/atomic"
)

var cid = atomic.Int32{}

type ClientState int

const (
	Closed ClientState = 1

	Open ClientState = 2

	Error ClientState = 3
)

type ClientControl struct {

	//Current client state.
	state chan ClientState
}

type ClientHandler interface {
	Close()

	Read()
}

// Client
// @Description: Define Client.
type Client struct {
	host string

	port int32

	id int32

	conn net.Conn

	ctx context.Context

	opts *cOptions

	handler []*ClientHandler

	session *smux.Session

	cct *ClientControl
}

// NewClient
//
//	@Description: Build a Client.
//	@param host
//	@param port
func NewClient(host string, port int32, cct *ClientControl) *Client {
	return &Client{
		host:    host,
		port:    port,
		ctx:     context.Background(),
		id:      cid.Add(1),
		cct:     cct,
		handler: make([]*ClientHandler, 0),
	}
}

func (c *Client) GetHost() string {
	return c.host
}

func (c *Client) GetPort() int32 {
	return c.port
}

func (c *Client) GetID() int32 {
	return c.id
}

func (c *Client) GetConn() net.Conn {
	return c.conn
}

func (c *Client) AddHandler(h ...*ClientHandler) {
	c.handler = append(c.handler, h...)
}

// Connection
//
//	@Description: connection to server.
//	@receiver c
//	@return error
func (c *Client) Connection(network string, option ...ClientOption) error {
	c.opts = clientOptions(option...)
	dialer := &net.Dialer{
		KeepAlive: c.opts.KeepAlive,
		Timeout:   c.opts.Timeout,
	}
	dial, err := dialer.DialContext(c.ctx, network, c.host+":"+fmt.Sprintf("%d", c.port))
	if err != nil {
		return err
	}
	c.cct.state <- Open
	c.conn = dial
	if c.opts.Smux != nil && c.opts.Smux.Enable {
		return c.handlerSmux()
	}
	return nil
}

func (c *Client) handlerSmux() error {
	config := smux.DefaultConfig()
	config.KeepAliveInterval = c.opts.Smux.Timeout
	config.KeepAliveTimeout = c.opts.Smux.Timeout
	config.KeepAliveDisabled = !c.opts.Smux.Enable
	session, err := smux.Client(c.conn, config)
	c.session = session
	if err != nil {
		log.Error("Open1 smux error %s", err.Error())
		return err
	}
	stream, err := session.Open()
	if err != nil {
		log.Error("Open session error:%s", err)
		return err
	}
	for {
		bytes := make([]byte, 4096)
		n, err := stream.Read(bytes)
		if err != nil {
			break
		} else {
			log.Info("Reader n %d", n)
		}
	}
	return nil
}

// Close
//
//	@Description: 关闭链接.
//	@receiver c
//	@return error
func (c *Client) Close() error {
	return c.conn.Close()
}
