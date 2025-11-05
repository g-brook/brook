/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	"github.com/brook/common/threading"
	"github.com/xtaci/smux"
)

var cid int32

//go:generate stringer -type=ClientState
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

func (b BaseClientHandler) Close(*ClientControl) {
}

func (b BaseClientHandler) Connection(*ClientControl) {}

func (b BaseClientHandler) Read(*exchange.Protocol, *ClientControl) error {
	return nil
}

func (b BaseClientHandler) Error(error, *ClientControl) {

}

func (b BaseClientHandler) Timeout(*ClientControl) {

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
// Return
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

func (c *Client) Connection(network string, option ...ClientOption) error {
	c.pre(network, option)
	threading.GoSafe(c.handleLoop)
	threading.GoSafe(c.readLoop)
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
	//OpenStream smux
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
		c.cct.Close()
		return err
	}
	c.session = session
	threading.GoSafe(c.sessionLoop)
	return nil
}

func (c *Client) setTimeout(dial net.Conn) {
	if c.opts.PingTime > 0 && !c.isSmux() {
		_ = dial.SetReadDeadline(time.Now().Add(c.opts.PingTime))
	}
}

func (c *Client) OpenTunnel(config *configs.ClientTunnelConfig) error {
	if !c.isSmux() {
		return nil
	}
	//copy.
	client := GetTunnelClient(config.TunnelType, config)
	if client == nil {
		log.Error("Not found [%s] tunnel client, Pleas check.", config.TunnelType)
		return errors.New("not found tunnel client")
	}
	err := client.Open(c.session)
	if err != nil {
		log.Error("Open tunnel error, close client:%v", config.TunnelType)
		c.cct.Close()
	}
	return err
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

// getAddress returns the complete address of the client in the format "host:port"
// It combines the host and port obtained from the client's methods into a single string
func (c *Client) getAddress() string {
	// Using fmt.Sprintf to format the host and port into a string with the format "host:port"
	return fmt.Sprintf("%s:%d", c.GetHost(), c.GetPort())
}

func (c *Client) isSmux() bool {
	return c.opts.Smux != nil
}

// IsConnection checks if the client connection is active and in the correct state
// This method is used to verify whether the client has an active connection
//
// Parameters:
//   - None
//
// Returns:
//   - bool: Returns true if connection exists and is in Active state, false otherwise
func (c *Client) IsConnection() bool {
	// Check if connection object is not nil and state is Active
	return c.conn != nil && c.state == Active
}

// handleLoop manages the client's connection lifecycle and event handling
func (c *Client) handleLoop() {
	// _close is a nested function to handle connection cleanup
	//Close connection.
	_close := func() error {
		// Check if connection is already closed
		if c.conn != nil {
			_ = c.conn.Close()
		}
		if c.session != nil {
			_ = c.session.Close()
		}
		return nil
	}
	// Main event loop handling various client events
	for {
		select {
		// Handle state changes
		case c.state = <-c.cct.state:
			log.Debug("Client state change,%d:%s", c.port, c.state.String())
			if c.state == Active {
				c.revReadNext()
				for _, t := range c.handlers {
					// Notify all handlers of connection establishment
					t.Connection(c.cct)
				}
			}
			if c.state == Closed {
				_ = _close()
				// Close connection and notify handlers of closure
				for _, t := range c.handlers {
					t.Close(c.cct)
				}
			}
		case err := <-c.cct.errors:
			// Handle errors
			//sendError.
			for _, t := range c.handlers {
				// Notify all handlers of errors
				t.Error(err, c.cct)
			}
		case b := <-c.cct.read:
			// Handle incoming data
			for _, t := range c.handlers {
				// Process received data through all handlers
				err := t.Read(b, c.cct)
				if err != nil {
					_ = c.error("Read error", err)
				}
			}
		case bytes := <-c.cct.write:
			// Handle outgoing data
			_, _ = c.rw.Write(bytes)
			// Write data to the connection
		case <-c.cct.timeout:
			// Handle timeout events
			for _, t := range c.handlers {
				// Notify all handlers of timeout
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
				c.cct.Close()
				return
			}
		}
	}
}

func (c *ClientControl) Write(bytes []byte) error {
	if c.cli.rw == nil {
		log.Warn("Connection closed")
		return errors.New("connection closed")
	}
	c.write <- bytes
	return nil
}

func (c *ClientControl) Close() {
	c.state <- Closed
}

func (c *ClientControl) Error(err error) {
	c.errors <- err
}
