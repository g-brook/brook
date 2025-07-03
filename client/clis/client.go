package clis

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/xtaci/smux"
)

var cid int32

type ClientState int

const (
	NotAction ClientState = 0

	Closed ClientState = 1

	Active ClientState = 2
)

type ClientControl struct {

	//Current client state.
	state chan ClientState

	read chan *exchange.Protocol

	errors chan error

	revRead chan struct{}

	timeout chan bool

	write chan []byte

	cli *Client

	list *list.List
}

type ClientHandler interface {
	//
	// Close
	//  @Description: Shutdown.
	//
	Close(cct *ClientControl)

	//
	// Connection
	//  @Description: Connection.
	//  @param cct
	//
	Connection(cct *ClientControl)

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

func (b BaseClientHandler) Connection(cct *ClientControl) {}

func (b BaseClientHandler) Read(buffer *exchange.Protocol, cct *ClientControl) error {
	return nil
}

func (b BaseClientHandler) Error(err error, cct *ClientControl) {

}

func (b BaseClientHandler) Timeout(cct *ClientControl) {

}

// Client
// @Description: Define Client.
type Client struct {
	host string

	port int

	id int32

	conn net.Conn

	opts *cOptions

	handlers []ClientHandler

	cct *ClientControl

	rw io.ReadWriter

	state ClientState

	network string

	session *smux.Session
}

// NewClient creates a new Client instance with the provided host and port.
// It initializes the client's context, ID, and control channels.
// Parameters:
//   - host: The server host address.
//   - port: The server port number.
//
// Retur
// Returns:
//   - *Client: A pointer to the newly created Client instance.
func NewClient(host string, port int) *Client {
	return &Client{
		host:  host,
		port:  port,
		id:    atomic.AddInt32(&cid, 1),
		state: NotAction,
		cct: &ClientControl{
			state:   make(chan ClientState),
			read:    make(chan *exchange.Protocol, 1024),
			errors:  make(chan error),
			timeout: make(chan bool),
			revRead: make(chan struct{}),
			write:   make(chan []byte, 1024),
			list:    list.New(),
		},
		handlers: make([]ClientHandler, 0),
	}
}

func (c *Client) GetHost() string {
	return c.host
}

func (c *Client) GetPort() int {
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

// Reconnection is
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
	c.pre(network, option)
	go c.handleLoop()
	go c.readLoop()
	err := c.doConnection()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) pre(network string, option []ClientOption) {
	c.network = network
	c.opts = clientOptions(option...)
	c.AddHandler(c.opts.handlers...)
	c.cct.cli = c
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
	timeout, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()
	if dial, err := dialer.DialContext(timeout, c.network, c.host+":"+fmt.Sprintf("%d", c.port)); err != nil {
		return c.error(
			fmt.Sprintf("Connection to %s:%d,error", c.host, c.port),
			err,
		)
	} else {
		c.setTimeout(dial)
		c.conn = dial
		c.cct.state <- Active
		c.cct.cli = c
	}
	c.rw = c.conn
	if !c.isSmux() {
		log.Info("👍---->Connection %s success OK.✅--->", c.getAddress())
		return nil
	}
	log.Info("👍---->Connection %s %s success OK.✅--->", c.getAddress(), "^ Tunnel ^")
	//open smux
	openSmux := func() (*smux.Session, error) {
		config := smux.DefaultConfig()
		config.KeepAliveDisabled = !c.opts.Smux.KeepAlive
		if session, err := smux.Client(c.conn, config); err != nil {
			return nil, c.error("New smux Client error", err)
		} else {
			return session, nil
		}
	}
	session, err := openSmux()
	if err != nil {
		log.Error("Active smux Client error %v", err)
		return err
	}
	c.session = session
	go c.sessionLoop()
	return nil
}

func (c *Client) setTimeout(dial net.Conn) {
	if c.opts.PingTime > 0 && !c.isSmux() {
		_ = dial.SetReadDeadline(time.Now().Add(c.opts.PingTime))
	}
}

// OpenTunnel
//
//	@Description: Active connection to
//	@receiver c
//	@param name
func (c *Client) OpenTunnel(config *configs.ClientTunnelConfig) error {
	if !c.isSmux() {
		return nil
	}
	//copy.
	client := GetTunnelClient(config.Type, config)
	if client == nil {
		log.Error("Not found [%s] tunnel client, Pleas check.", config.Type)
		return errors.New("not found tunnel client")
	}
	return client.Open(c.session)
}

func (c *Client) error(str string, err error) error {
	if err == nil {
		err = errors.New(str)
	}
	log.Error("%s %s", str, err.Error())
	return err
}

func (c *Client) readLoop() {
	if c.isSmux() {
		return
	}
	<-c.cct.revRead
	clientFunction := func() error {
		protocol, err := exchange.Decoder(c.rw)
		if err != nil {
			if err == io.EOF {
				_ = c.error("Close connection:"+c.getAddress(), err)
				c.cct.state <- Closed
				return err
			} else {
				var opErr *net.OpError
				if errors.As(err, &opErr) && opErr.Timeout() {
					c.setTimeout(c.conn)
					c.cct.timeout <- true
				}
			}
			return nil
		}
		c.cct.read <- protocol
		return nil
	}
	for {
		err := clientFunction()
		if err == io.EOF {
			<-c.cct.revRead
		}
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
//	@receiver c
func (c *Client) IsConnection() bool {
	return c.conn != nil && c.state == Active
}

func (c *Client) handleLoop() {
	//Close connection.
	_close := func() error {
		//closed.
		if c.state != Closed && c.conn != nil {
			return c.conn.Close()
		}
		return nil
	}
	for {
		select {
		case c.state = <-c.cct.state:
			log.Debug("Client state change:%d", c.state)
			if c.state == Active {
				c.revReadNext()
				for _, t := range c.handlers {
					t.Connection(c.cct)
				}
			}
			if c.state == Closed {
				_ = _close()
				for _, t := range c.handlers {
					t.Close(c.cct)
				}
			}
		case err := <-c.cct.errors:
			//sendError.
			for _, t := range c.handlers {
				t.Error(err, c.cct)
			}
		case b := <-c.cct.read:
			for _, t := range c.handlers {
				err := t.Read(b, c.cct)
				if err != nil {
					_ = c.error("Read error", err)
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

func (c *Client) revReadNext() {
	select {
	case c.cct.revRead <- struct{}{}:
	default:
	}
}

func (c *Client) sessionLoop() {
	if c.session != nil {
		for {
			select {
			case <-c.session.CloseChan():
				log.Warn("Tunnel Session closed %v", c.session.RemoteAddr())
				c.cct.state <- Closed
				return
			}
		}
	}
}

// Write
//
//	@Description: Write data..
//	@receiver c
//	@param bytes
func (c *ClientControl) Write(bytes []byte) error {
	if c.cli.rw == nil {
		log.Warn("Connection closed")
		return errors.New("connection closed")
	}
	c.write <- bytes
	return nil
}

// Close
//
//	@Description: Close Client.
//	@receiver c
func (c *ClientControl) Close() {
	c.state <- Closed
}

// Error
//
//	@Description: Print error.
//	@receiver c
//	@param err
func (c *ClientControl) Error(err error) {
	c.errors <- err
}
