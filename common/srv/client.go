package srv

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/xtaci/smux"
	"io"
	"net"
	"sync/atomic"
	"time"
)

var cid int32

type ClientState int

const (
	Closed ClientState = 1

	Open ClientState = 2

	Error ClientState = 3
)

type ClientControl struct {
	//Current client state.
	state chan ClientState

	read chan *exchange.Protocol

	errors chan error

	close chan bool

	timeout chan bool

	write chan []byte

	cli *Client

	list *list.List
}

type ClientHandler interface {
	//
	// Close
	//  @Description: Close.
	//
	Close(cct *ClientControl)

	//
	// Read
	//  @Description: bytes cli
	//  @param bytes
	//  @param cli
	//  @return int
	//  @return error
	//
	Read(buffer *exchange.Protocol, cct *ClientControl) error

	//
	// Error
	//  @Description: ERROR
	//  @param err
	//
	Error(err error, cct *ClientControl)

	//
	// Timeout
	//  @Description:
	//  @param cct
	//
	Timeout(cct *ClientControl)
}

type BaseClientHandler struct {
}

func (b BaseClientHandler) Close(cct *ClientControl) {

}

func (b BaseClientHandler) Read(buffer *exchange.Protocol, cct *ClientControl) (int, error) {
	return 0, nil
}

func (b BaseClientHandler) Error(err error, cct *ClientControl) {

}

func (b BaseClientHandler) Timeout(cct *ClientControl) {

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

	handlers []ClientHandler

	cct *ClientControl

	rw io.ReadWriter

	state ClientState

	network string
}

// NewClient
//
//	@Description: Build a Client.
//	@param host
//	@param port
func NewClient(host string, port int32) *Client {
	return &Client{
		host: host,
		port: port,
		ctx:  context.Background(),
		id:   atomic.AddInt32(&cid, 1),
		cct: &ClientControl{
			state:   make(chan ClientState, 1),
			close:   make(chan bool, 1),
			read:    make(chan *exchange.Protocol, 1000),
			errors:  make(chan error),
			timeout: make(chan bool, 1),
			write:   make(chan []byte, 1000),
			list:    list.New(),
		},
		handlers: make([]ClientHandler, 0),
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

func (c *Client) AddHandler(h ...ClientHandler) {
	c.handlers = append(c.handlers, h...)
}

func (c *Client) Reconnection() error {
	if !c.IsConnection() {
		return c.doConnection()
	}
	return nil
}

// Connection
//
//	@Description: connection to server.
//	@receiver c
//	@return error
func (c *Client) Connection(network string, option ...ClientOption) error {
	c.network = network
	c.opts = clientOptions(option...)
	c.AddHandler(c.opts.handlers...)
	go c.handleLoop()
	go c.readLoop()
	err := c.doConnection()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doConnection() error {
	if c.conn != nil {
		c.conn = nil
		c.rw = nil
	}
	dialer := &net.Dialer{
		KeepAlive: c.opts.KeepAlive,
		Timeout:   c.opts.Timeout,
	}
	if dial, err := dialer.DialContext(c.ctx, c.network, c.host+":"+fmt.Sprintf("%d", c.port)); err != nil {
		return c.error(
			fmt.Sprintf("Connection to %s:%d,error", c.host, c.port),
			err,
		)
	} else {
		c.setTimeout(dial)
		c.conn = dial
		c.cct.state <- Open
		c.cct.cli = c
		if c.isSmux() {
			log.Info("👍---->Connection %s %s success OK.✅--->", c.getAddress(), "^ tunnel ^")
		} else {
			log.Info("👍---->Connection %s success OK.✅--->", c.getAddress())
		}
	}
	c.rw = c.conn
	return nil
}

func (c *Client) setTimeout(dial net.Conn) {
	if c.opts.PingTime > 0 && !c.isSmux() {
		_ = dial.SetReadDeadline(time.Now().Add(c.opts.PingTime))
	}
}

// OpenTunnel
//
//	@Description: Open connection to
//	@receiver c
//	@param name
func (c *Client) OpenTunnel(name string) error {
	if !c.isSmux() {
		return nil
	}
	openSmux := func() (*smux.Session, error) {
		config := smux.DefaultConfig()
		config.KeepAliveInterval = c.opts.Smux.Timeout
		config.KeepAliveTimeout = c.opts.Smux.Timeout
		config.KeepAliveDisabled = !c.opts.Smux.Enable
		if session, err := smux.Client(c.conn, config); err != nil {
			return nil, c.error("New smux Client error", err)
		} else {
			return session, nil
		}
	}
	session, err := openSmux()
	if err != nil {
		return err
	}
	//copy.
	client := GetTunnelClient(name)
	if client == nil {
		log.Error("Not found [%s] tunnel client, Pleas check.", name)
		return errors.New("not found tunnel client")
	}
	go client.Open(session)
	return nil
}

func (c *Client) error(str string, err error) error {
	if err == nil {
		err = errors.New(str)
	}
	log.Error("%s %s", str, err.Error())
	c.cct.state <- Error
	return err
}

func (c *Client) readLoop() {
	if c.isSmux() {
		return
	}
	// Is not Tunnel connection.
	clientFunction := func() {
		protocol, err := exchange.Decoder(c.rw)
		if err != nil {
			if err == io.EOF {
				_ = c.error("Close connection:"+c.getAddress(), err)
				c.cct.state <- Closed
				c.cct.close <- true
			} else {
				var opErr *net.OpError
				if errors.As(err, &opErr) && opErr.Timeout() {
					c.setTimeout(c.conn)
					c.cct.timeout <- true
				}
			}

		} else {
			c.cct.read <- protocol
		}
	}
	for {
		if c.rw == nil || c.conn == nil {
			continue
		}
		clientFunction()
	}
}

func (c *Client) getAddress() string {
	return fmt.Sprintf("%s:%d", c.GetHost(), c.GetPort())
}

func (c *Client) isSmux() bool {
	return c.opts.Smux != nil
}

// IsConnection
//
//	@Description: 是否存在连接.
//	@receiver c
func (c *Client) IsConnection() bool {
	return c.conn != nil && c.state == Open
}

func (c *Client) chState() {
	c.state = <-c.cct.state
}

func (c *Client) handleLoop() {
	//Close connection.
	_close := func() error {
		//closed.
		if c.state != Closed {
			return c.conn.Close()
		}
		return nil
	}
	for {
		select {
		case c.state = <-c.cct.state:
			log.Info("Client state change:%d", c.state)
		case <-c.cct.close:
			for _, t := range c.handlers {
				t.Close(c.cct)
			}
			_ = _close()
		case err := <-c.cct.errors:
			for _, t := range c.handlers {
				t.Error(err, c.cct)
			}
			//sendError.
			c.cct.state <- Error
		case b := <-c.cct.read:
			for _, t := range c.handlers {
				err := t.Read(b, c.cct)
				if err != nil {
					_ = c.error("Read error", err)
					break
				}
			}
		case bytes := <-c.cct.write:
			_, _ = c.rw.Write(bytes)
		case <-c.cct.timeout:
			for _, t := range c.handlers {
				t.Timeout(c.cct)
			}
		}
	}
}

// Write
//
//	@Description: Write data..
//	@receiver c
//	@param bytes
func (c *ClientControl) Write(bytes []byte) {
	c.write <- bytes
}

// Close
//
//	@Description: Close Client.
//	@receiver c
func (c *ClientControl) Close() {
	c.close <- true
}

// Error
//
//	@Description: Print error.
//	@receiver c
//	@param err
func (c *ClientControl) Error(err error) {
	c.errors <- err
}
