package clis

import (
	"github.com/brook/client/cli"
	"github.com/brook/common/exchange"
	"time"
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
	transport *Transport
	commands  map[exchange.Cmd]CmdNotify
	UnId      string
}

func (b *managerTransport) Close(_ *ClientControl) {
	cli.UpdateStatus("offline")
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
	message, ok := b.commands[r.Cmd]
	if ok {
		return message(r)
	}
	return nil
}

// NewManagerTransport This function creates a new managerTransport object and returns it
func NewManagerTransport(tr *Transport) *managerTransport {
	// Create a new managerTransport object
	transport := &managerTransport{
		// Set the transport field of the managerTransport object to the given Transport object
		transport: tr,
		commands:  make(map[exchange.Cmd]CmdNotify),
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

func (b *managerTransport) BindUnId(unId string) {
	b.UnId = unId
}

func (b *managerTransport) AddMessage(cmd exchange.Cmd, notify CmdNotify) {
	b.commands[cmd] = notify
}
