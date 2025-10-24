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

package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	. "github.com/brook/common/transport"
)

var (
	PHttpErr = errors.New("the server requested http, but the request was https")

	PHttpsErr = errors.New("the server requested https, but the request was http")

	PTimeout = errors.New("read timeout")

	httpMethods = [][]byte{
		[]byte(http.MethodGet),
		[]byte(http.MethodPost),
		[]byte(http.MethodConnect),
		[]byte(http.MethodDelete),
		[]byte(http.MethodHead),
		[]byte(http.MethodOptions),
		[]byte(http.MethodPatch),
		[]byte(http.MethodPut),
		[]byte(http.MethodTrace)}
)

type HttpConn struct {
	ch        Channel
	buf       bytes.Buffer
	dataCh    chan struct{}
	mu        sync.Mutex
	https     bool
	handshake bool
	closed    chan struct{}
}

/**
 * Creates a new HTTP connection with the given channel and HTTPS flag
 * @param ch The channel for the connection
 * @param https Whether to use HTTPS or not
 * @return A new HttpConn instance
 */
func newHttpConn(ch Channel, https bool) *HttpConn {
	// Create a new HttpConn instance with the provided channel and HTTPS flag
	// Initialize dataCh and closed channels for synchronization
	conn := &HttpConn{
		ch:     ch,                  // Channel for the connection
		dataCh: make(chan struct{}), // Channel for data synchronization
		closed: make(chan struct{}), // Channel to track connection closure
		https:  https,               // HTTPS flag indicating whether to use HTTPS
	}
	return conn // Return the newly created HttpConn instance
}

// Is Tls.
// isTLSHandshake checks if the given data represents a TLS handshake message
// It verifies that the data is at least 3 bytes long and that the first two bytes
// match the TLS handshake record type (0x16) and the major version number (0x03)
//
// Parameters:
//
//	data: A byte slice containing the data to be checked
//
// Returns:
//
//	bool: True if the data is a TLS handshake message, false otherwise
func isTLSHandshake(data []byte) bool {
	return len(data) >= 3 && data[0] == 0x16 && data[1] == 0x03
}

// isHTTPRequest checks if the given data represents an HTTP request
// by verifying if the first three characters are uppercase letters
func isHTTPRequest(data []byte) bool {
	//  Check if the data length is less than 3 bytes
	// If so, it can't be a valid HTTP request
	if len(data) < 3 {
		return false
	}

	for _, m := range httpMethods {
		if len(data) >= len(m) && bytes.Equal(data[:len(m)], m) {
			return true
		}
	}
	return false
}

// OnData handles incoming data for the HTTP connection
// It writes the data to an internal buffer and signals that new data is available
func (h *HttpConn) OnData(b []byte) {
	h.mu.Lock()    // Acquire lock to ensure thread-safe buffer writing
	h.buf.Write(b) // Append received data to the internal buffer
	h.mu.Unlock()  // Release the lock
	// Non-blocking signal that new data is available
	// If the channel is already full, this does nothing (default case)
	select {
	case h.dataCh <- struct{}{}: // Send empty struct to signal new data
	default: // If channel is full, do nothing
	}
}

// Read implements the io.Reader interface for HttpConn.
// It reads data from the connection buffer with proper synchronization and protocol validation.
func (h *HttpConn) Read(b []byte) (n int, err error) {
	// Infinite loop to continuously attempt reading data
	for {
		// Lock the mutex to ensure thread-safe access to shared resources
		h.mu.Lock()
		// Check if there's data available in the internal buffer
		if h.buf.Len() > 0 {
			// Read available data into the provided byte slice
			read, _ := h.buf.Read(b)
			// Unlock the mutex after reading from buffer
			h.mu.Unlock()
			// Protocol validation for HTTPS connection
			if h.https && !h.handshake {
				// Verify if the data is part of TLS handshake
				if !isTLSHandshake(b[:read]) {
					return 0, PHttpsErr
				}
				// Mark handshake as completed after successful validation
				h.handshake = true
			} else if !h.https && !isHTTPRequest(b) {
				// Validate for HTTP protocol if not HTTPS
				return 0, PHttpErr
			}
			// Return the number of bytes read and no error
			return read, nil
		}
		// Unlock the mutex if no data was available in buffer
		h.mu.Unlock()
		// Wait for data with multiple exit conditions
		select {
		// Case when new data arrives through the data channel
		case <-h.dataCh:
			// Case when timeout occurs after 30 seconds
		case <-time.After(30 * time.Second):
			return 0, PTimeout
			// Case when the connection is closed
		case <-h.closed:
			return 0, io.EOF
		}
	}
}

func (h *HttpConn) Write(b []byte) (n int, err error) {
	//That's a hack, but we don't want to write to the underlying connection
	//Do not use the Write method of the connection
	//return h.ch.GetWriter().Write(b)
	return h.ch.Write(b)
}

func (h *HttpConn) Close() error {
	select {
	case <-h.closed:
	default:
		close(h.closed)
	}
	return h.ch.Close()
}

func (h *HttpConn) LocalAddr() net.Addr {
	return h.ch.LocalAddr()
}

func (h *HttpConn) RemoteAddr() net.Addr {
	return h.ch.RemoteAddr()
}

func (h *HttpConn) SetDeadline(t time.Time) error {
	return h.ch.SetDeadline(t)
}

func (h *HttpConn) SetReadDeadline(t time.Time) error {
	return h.ch.SetReadDeadline(t)
}

func (h *HttpConn) SetWriteDeadline(t time.Time) error {
	return h.ch.SetWriteDeadline(t)
}

type responseWriter struct {
	conn     net.Conn
	httpConn *HttpConn
	header   http.Header
	wrote    bool
	status   int
	req      *http.Request
	body     *bytes.Buffer
}

func newResponseWriter(conn net.Conn, httpConn *HttpConn, req *http.Request) *responseWriter {
	return &responseWriter{
		conn: conn, httpConn: httpConn, header: make(http.Header), req: req, body: bytes.NewBuffer(make([]byte, 0)),
	}
}

func (r *responseWriter) Write(bt []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(bt)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if r.wrote {
		return
	}
	r.wrote = true
	r.status = statusCode
}

// finish completes the response writing process
// It ensures proper header setup and writes the complete response to the connection
func (r *responseWriter) finish(err error) {
	// Check if headers have been written, if not write default status 200
	if !r.wrote {
		if err != nil {
			r.WriteHeader(http.StatusBadRequest)
		} else {
			r.WriteHeader(http.StatusOK)
		}

	}
	// Set Content-Length header if not already set
	if r.header.Get("Content-Length") == "" {
		r.header.Set("Content-Length", strconv.Itoa(r.body.Len()))
	}
	// Create a new buffer to build the response
	resp := bytes.NewBuffer(make([]byte, 0))
	// Check if using HTTP/1.1 or later
	is11 := r.req.ProtoAtLeast(1, 1)
	// Write the status line to the response buffer
	writeHeaderLine(resp, is11, r.status)
	// Write all headers to the response buffer
	for k, v := range r.header {
		for _, s := range v {
			_, _ = fmt.Fprintf(resp, "%s: %s\r\n", k, s)
		}
	}
	// Write the final empty line to signify end of headers
	_, _ = fmt.Fprintf(resp, "\r\n")
	// Write the body content and reset the body buffer
	resp.Write(r.body.Bytes())
	r.body.Reset()
	if err != nil && r.httpConn != nil {
		_, _ = r.httpConn.Write(resp.Bytes())
		return
	}
	// Write the complete response to the connection
	_, _ = r.conn.Write(resp.Bytes())
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) error(err error) {
	if err == nil {
		return
	}
	if !errors.Is(err, PHttpErr) && !errors.Is(err, PHttpsErr) {
		return
	}
	r.req = &http.Request{}
	r.WriteHeader(http.StatusBadRequest)
	_, _ = r.Write([]byte(err.Error()))
	r.finish(err)
}

func writeHeaderLine(rw *bytes.Buffer, is11 bool, code int) {
	if is11 {
		_, _ = rw.WriteString("HTTP/1.1 ")
	} else {
		_, _ = rw.WriteString("HTTP/1.0 ")
	}
	if text := http.StatusText(code); text != "" {
		_, _ = rw.WriteString(fmt.Sprintf("%d %s", code, text))
		_, _ = rw.WriteString("\r\n")
	} else {
		_, _ = rw.WriteString(fmt.Sprintf("%03d status code %d\r\n", code, code))
	}
}
