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

// Package transport provides network transport abstractions
// This file defines the Channel interface which represents a network communication channel
package transport

import (
	"io"
	"net"

	"github.com/brook/common"
)

type CloseEvent func(channel Channel)

// Channel interface represents a network communication channel
// It combines io.Reader and io.Writer with net.Conn to provide a comprehensive interface
// for network operations. This interface is designed to be implemented by different
// transport protocols while providing a consistent API for upper layers.
type Channel interface {
	io.Reader

	io.Writer

	// Conn net.Conn provides basic network connection functionality
	net.Conn

	// GetReader returns the reader part of the channel
	// This can be used to get a specific reader implementation
	GetReader() io.Reader

	// GetWriter returns the writer part of the channel
	// This can be used to get a specific writer implementation
	GetWriter() io.Writer

	// RemoteAddr returns the remote network address
	// This overrides the method from net.Conn to provide
	// a more specific implementation for this transport
	RemoteAddr() net.Addr

	// LocalAddr returns the local network address
	// This overrides the method from net.Conn to provide
	// a more specific implementation for this transport
	LocalAddr() net.Addr

	// GetConn returns the underlying network connection
	// This provides access to the raw network connection
	// for cases where direct access is needed
	GetConn() net.Conn

	// GetId returns the unique identifier for the channel
	GetId() string

	// Done returns a channel that is closed when the channel is closed
	// This can be used to wait for the channel to be closed
	//Close
	Done() <-chan struct{}

	// IsClose isClose.
	IsClose() bool

	SendTo(by []byte, addr net.Addr) (int, error)
	//
	// GetAttr
	//  @Description: getKey.
	//  @param key
	//
	GetAttr(key common.KeyType) (interface{}, bool)

	// OnClose CloseEvent The Channel Setter CloseEvent
	OnClose(event CloseEvent)
}
