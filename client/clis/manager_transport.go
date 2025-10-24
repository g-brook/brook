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

	"github.com/brook/client/cli"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
)

var ManagerTransport *managerTransport

type CmdNotify func(*exchange.Protocol) error

// InitManagerTransport This function initializes the ManagerTransport with a given transport
func InitManagerTransport(transport *Transport) {
	// Create a new ManagerTransport with the given transport
	ManagerTransport = NewManagerTransport(transport)
}

type managerTransport struct {
	BaseClientHandler
	transport       *Transport
	tunnelTransport *Transport
	commands        map[exchange.Cmd]CmdNotify
	UnId            string
	configs         map[string]*configs.ClientTunnelConfig
}

func (b *managerTransport) WithTunnelTransport(t *Transport) {
	b.tunnelTransport = t
}

func (b *managerTransport) Close(_ *ClientControl) {
	cli.UpdateStatus("offline")
	if b.tunnelTransport != nil {
		b.tunnelTransport.Close()
	}
}

func (b *managerTransport) Connection(_ *ClientControl) {
	cli.UpdateStatus("online")
}

func (b *managerTransport) Read(r *exchange.Protocol, cct *ClientControl) error {
	//Heart info.
	if r.Cmd == exchange.Heart {
		t, _ := exchange.Parse[exchange.Heartbeat](r.Data)
		startTime := t.StartTime
		endTime := time.Now().UnixMilli()
		cli.UpdateSpell(endTime - startTime)
		return nil
	}
	return b.PushMessage(r)
}

// NewManagerTransport This function creates a new managerTransport object and returns it
func NewManagerTransport(tr *Transport) *managerTransport {
	// Create a new managerTransport object
	transport := &managerTransport{
		// Set the transport field of the managerTransport object to the given Transport object
		transport: tr,
		commands:  make(map[exchange.Cmd]CmdNotify),
		configs:   make(map[string]*configs.ClientTunnelConfig),
	}
	// Return the new managerTransport object
	return transport
}

// GetTransport This function returns the transport associated with the receiver
func (b *managerTransport) GetTransport() *Transport {
	// Return the transport associated with the b
	return b.transport
}

// SyncWrite This function is a method of the managerTransport struct and is used to synchronously write a message to the transport with a specified timeout.
func (b *managerTransport) SyncWrite(message exchange.InBound, timeout time.Duration) (*exchange.Protocol, error) {
	// Call the SyncWrite method of the transport struct and pass in the message and timeout
	return b.transport.SyncWrite(
		message,
		timeout,
	)
}

func (b *managerTransport) GetConfig(proxyId string) *configs.ClientTunnelConfig {
	return b.configs[proxyId]
}

func (b *managerTransport) PutConfig(config *configs.ClientTunnelConfig) {
	b.configs[config.ProxyId] = config
}

func (b *managerTransport) BindUnId(unId string) {
	b.UnId = unId
}

func (b *managerTransport) AddMessageNotify(cmd exchange.Cmd, notify CmdNotify) {
	b.commands[cmd] = notify
}

func (b *managerTransport) PushMessage(r *exchange.Protocol) error {
	message, ok := b.commands[r.Cmd]
	if ok {
		return message(r)
	}
	return nil
}
