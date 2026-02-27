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
	"time"

	"github.com/g-brook/brook/client/cli"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/log"
)

// Transport
// @Description:Transport manages client and request tracking.
type Transport struct {

	// client　is net connection.
	client *Client

	host string

	port int

	config *configs.ClientConfig

	reconnect *ReconnectManager
}

// NewTransport
//
//	@Description: Init Transport.
//	@param ct
//	@return Transport
func NewTransport(config *configs.ClientConfig) *Transport {
	//start reconnection.
	return &Transport{
		host:      config.ServerHost,
		port:      config.ServerPort,
		config:    config,
		reconnect: NewReconnectionManager(time.Second * 5),
	}
}

// Connection establishes a new connection with the server using the provided options
// It sets up a new client, adds a handler for checking the connection,
// and either adds the transport to a reconnection list or opens a tunnel based on the connection result
func (t *Transport) Connection(opts ...ClientOption) {
	// Create a new client with the specified host and port
	t.client = NewClient(t.host, t.port)
	// Attempt to establish a TCP connection with the provided options
	err := t.client.Connection("tcp", opts...)
	// Add a CheckHandler to manage the connection state
	t.client.AddHandler(&CheckHandler{
		transport: t,
	})
	//The error add to reconnection list.
	if err != nil {
		// If connection fails, log a warning and add this transport to a checking list for reconnection
		log.Warn("Connection to server error:%s", err)
		addChecking(t)
	} else {
		// If connection is successful, open a tunnel for data transmission
		t.openTunnel()
	}
}

// Close closes the transport by closing the underlying client connection.
// It ensures proper cleanup of resources associated with the transport.
func (t *Transport) Close() {
	// Close the client connection using the client's connection table (cct)
	t.client.cct.Close()
	if t.client.isSmux() {
		cli.UpdateConnState(true)
	}
}

func (t *Transport) openTunnel() {
	if t.client.isSmux() && t.config.Tunnels != nil {
		for _, cfg := range t.config.Tunnels {
			if err := t.client.OpenTunnel(cfg); err != nil {
				log.Warn("Connection to server error:%s %v", cfg.TunnelType, err)
			}
		}
	}
}

func (t *Transport) SyncWrite(message exchange.InBound, timeout time.Duration) (*exchange.Protocol, error) {
	return exchange.SyncWriteInBound(message, timeout, func(protocol *exchange.Protocol) error {
		return t.client.cct.Write(protocol.Bytes())
	})
}

type ClientScheduler struct {
}

func (t *ClientScheduler) Next(t2 time.Time) time.Time {
	return t2.Add(3 * time.Second)
}

type CheckHandler struct {
	BaseClientHandler
	transport *Transport
}

func (b *CheckHandler) Close(*ClientControl) {
	addChecking(b.transport)
}

func (b *CheckHandler) Read(r *exchange.Protocol, cct *ClientControl) error {
	//Heart info.
	if r.Cmd == exchange.Heart {
		log.Debug("Receiver PONG info: %v", cct.cli.getAddress())
		return nil
	}
	exchange.Tracker.Complete(r)
	return nil
}

func (b *CheckHandler) Timeout(cct *ClientControl) {
	var h = &exchange.Heartbeat{
		Value:     "PING",
		StartTime: time.Now().UnixMilli(),
	}
	request, _ := exchange.NewRequest(h)
	_ = cct.Write(request.Bytes())
}

func addChecking(tp *Transport) {
	reconnect := func() bool {
		client := tp.client
		if !client.IsConnection() {
			log.Warn("Connection %s Not Active, start reconnection.", client.getAddress())
			err := client.doConnection()
			if err != nil {
				log.Warn("Reconnection %s Fail, next time still running.", client.getAddress())
			} else {
				log.Info("👍<--Reconnection %s success OK.✅-->", client.getAddress())
				tp.openTunnel()
			}
		}
		return client.IsConnection()
	}
	tp.reconnect.TryReconnect(reconnect)
}
