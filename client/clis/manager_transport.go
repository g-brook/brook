package clis

import (
	"github.com/brook/common/exchange"
	"time"
)

var ManagerTransport *managerTransport

// InitManagerTransport This function initializes the ManagerTransport with a given transport
func InitManagerTransport(transport *Transport) {
	// Create a new ManagerTransport with the given transport
	ManagerTransport = NewManagerTransport(transport)
}

type managerTransport struct {
	transport *Transport
}

// NewManagerTransport This function creates a new managerTransport object and returns it
func NewManagerTransport(tr *Transport) *managerTransport {
	// Create a new managerTransport object
	transport := &managerTransport{
		// Set the transport field of the managerTransport object to the given Transport object
		transport: tr,
	}
	// Return the new managerTransport object
	return transport
}

// GetTransport This function returns the transport associated with the receiver
func (receiver *managerTransport) GetTransport() *Transport {
	// Return the transport associated with the receiver
	return receiver.transport
}

// SyncWrite This function is a method of the managerTransport struct and is used to synchronously write a message to the transport with a specified timeout.
func (receiver *managerTransport) SyncWrite(message exchange.InBound, timeout time.Duration) (*exchange.Protocol, error) {
	// Call the SyncWrite method of the transport struct and pass in the message and timeout
	return receiver.transport.SyncWrite(
		message,
		timeout,
	)
}
