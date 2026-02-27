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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/g-brook/brook/common/ringbuffer"
	. "github.com/g-brook/brook/common/transport"
)

var (
	PHttpErr = errors.New("the server requested http, but the request was https")

	PHttpsErr = errors.New("the server requested https, but the request was http")

	PTimeout = errors.New("read client timeout, connect close")

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

const (
	timeout = 30 * time.Second
)

type Conn struct {
	ch          Channel
	buffer      *ringbuffer.RingBuffer
	https       bool
	handshake   bool
	closed      chan struct{}
	dataCh      chan struct{}
	isWebSocket bool
	timeStop    *time.Timer
	closeOnce   sync.Once
}

/**
 * Creates a new HTTP connection with the given channel and HTTPS flag
 * @param ch The channel for the connection
 * @param https Whether to use HTTPS or not
 * @return A new HttpConn instance
 */
func newHttpConn(ch Channel, https bool) *Conn {
	// Create a new HttpConn instance with the provided channel and HTTPS flag
	// Initialize dataCh and closed channels for synchronization
	conn := &Conn{
		ch:       ch,                              // Channel for the connection
		buffer:   ringbuffer.NewRingBuffer(65535), // Buffer for incoming data
		dataCh:   make(chan struct{}, 1),          // Channel for data synchronization
		closed:   make(chan struct{}),             // Channel to track connection closure
		https:    https,                           // HTTPS flag indicating whether to use HTTPS
		timeStop: time.NewTimer(timeout),
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

// isHttpRequest checks if the given data represents an HTTP request
// by verifying if the first three characters are uppercase letters
func isHttpRequest(data []byte) bool {
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
func (h *Conn) OnData(b []byte) {
	n, err := h.buffer.Write(b)
	if err != nil {
		return
	}
	if n < len(b) {
		return
	}
	select {
	case h.dataCh <- struct{}{}:
	default:
	}
}

// Read implements the io.Reader interface for HttpConn.
// It reads data from the connection buffer with proper synchronization and protocol validation.
func (h *Conn) Read(b []byte) (n int, err error) {
	for {
		if !h.buffer.IsEmpty() {
			read, _ := h.buffer.Read(b)
			if h.isWebSocket {
				return read, nil
			}
			if h.https && !h.handshake {
				if !isTLSHandshake(b[:read]) {
					return 0, PHttpsErr
				}
				h.handshake = true
			} else if !h.https && !isHttpRequest(b) {
				// Validate for HTTP protocol if not HTTPS
				return 0, PHttpErr
			}
			if !h.timeStop.Stop() {
				<-h.timeStop.C
			}
			h.timeStop.Reset(timeout)
			return read, nil
		}
		select {
		case <-h.dataCh:
		case <-h.timeStop.C:
			if h.isWebSocket {
				return 0, nil
			}
			return 0, PTimeout
		case <-h.closed:
			return 0, io.EOF
		}
	}
}

func (h *Conn) Write(b []byte) (n int, err error) {
	//That's a hack, but we don't want to write to the underlying connection
	//Do not use the Write method of the connection
	//return h.ch.GetWriter().Write(b)
	return h.ch.Write(b)
}

func (h *Conn) Close() error {
	h.closeOnce.Do(func() {
		select {
		case <-h.closed:
		default:
			close(h.closed)
		}
		h.timeStop.Stop()
		_ = h.ch.Close()
	})
	return nil
}

func (h *Conn) LocalAddr() net.Addr {
	return h.ch.LocalAddr()
}

func (h *Conn) RemoteAddr() net.Addr {
	return h.ch.RemoteAddr()
}

func (h *Conn) SetDeadline(t time.Time) error {
	return h.ch.SetDeadline(t)
}

func (h *Conn) SetReadDeadline(t time.Time) error {
	return h.ch.SetReadDeadline(t)
}

func (h *Conn) SetWriteDeadline(t time.Time) error {
	return h.ch.SetWriteDeadline(t)
}

type responseWriter struct {
	conn     net.Conn
	httpConn *Conn
	header   http.Header
	wrote    bool
	status   int
	req      *http.Request
	body     *bytes.Buffer
	timeStop time.Time
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.httpConn.isWebSocket = true
	return r.conn, bufio.NewReadWriter(bufio.NewReader(r.conn), bufio.NewWriter(r.conn)), nil
}

func newResponseWriter(conn net.Conn,
	httpConn *Conn,
	req *http.Request) *responseWriter {
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
func (r *responseWriter) finish(err error, req *http.Request) error {
	// Check if headers have been written, if not write default status 200
	if !r.wrote {
		if err != nil {
			r.WriteHeader(http.StatusBadRequest)
			r.header.Set("Connection", "close")
		} else {
			r.WriteHeader(http.StatusOK)
			if req != nil {
				r.header.Set("Connection", req.Header.Get("Connection"))
			} else {
				r.header.Set("Connection", "keep-alive")
			}
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
		return err
	}
	// Write the complete response to the connection
	_, err = r.conn.Write(resp.Bytes())
	return err
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
	_ = r.finish(err, nil)
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
