package remote

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
)

var cid = atomic.Int32{}

type ClientHandler interface {
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
}

// NewClient
//
//	@Description: Build a Client.
//	@param host
//	@param port
func NewClient(host string, port int32) *Client {
	return &Client{
		host:    host,
		port:    port,
		ctx:     context.Background(),
		id:      cid.Add(1),
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
	c.conn = dial
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
