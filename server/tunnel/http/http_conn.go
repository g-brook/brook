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

var httpMethods = [][]byte{
	[]byte(http.MethodGet),
	[]byte(http.MethodPost),
	[]byte(http.MethodConnect),
	[]byte(http.MethodDelete),
	[]byte(http.MethodHead),
	[]byte(http.MethodOptions),
	[]byte(http.MethodPatch),
	[]byte(http.MethodPut),
	[]byte(http.MethodTrace)}

var (
	PHttpErr = errors.New("the server requested http, but the request was https")

	PHttpsErr = errors.New("the server requested https, but the request was http")

	PTimeout = errors.New("read timeout")
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

func newHttpConn(ch Channel, https bool) *HttpConn {
	conn := &HttpConn{
		ch:     ch,
		dataCh: make(chan struct{}),
		closed: make(chan struct{}),
		https:  https,
	}
	return conn
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

func (h *HttpConn) OnData(b []byte) {
	h.mu.Lock()
	h.buf.Write(b)
	h.mu.Unlock()
	select {
	case h.dataCh <- struct{}{}:
	default:
	}
}

func (h *HttpConn) Read(b []byte) (n int, err error) {
	for {
		h.mu.Lock()
		if h.buf.Len() > 0 {
			read, _ := h.buf.Read(b)
			h.mu.Unlock()
			if h.https && !h.handshake {
				if !isTLSHandshake(b[:read]) {
					return 0, PHttpsErr
				}
				h.handshake = true
			} else if !h.https && !isHTTPRequest(b) {
				return 0, PHttpErr
			}
			return read, nil
		}
		h.mu.Unlock()
		select {
		case <-h.dataCh:
		case <-time.After(30 * time.Second):
			return 0, PTimeout
		case <-h.closed:
			return 0, io.EOF
		}
	}
}

func (h *HttpConn) Write(b []byte) (n int, err error) {
	//That's a hack, but we don't want to write to the underlying connection
	//Do not use the Write method of the connection
	return h.ch.GetWriter().Write(b)
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

func (h *HttpConn) writeErr() {
}

type responseWriter struct {
	conn   net.Conn
	header http.Header
	wrote  bool
	status int
	req    *http.Request
	body   *bytes.Buffer
}

func newResponseWriter(conn net.Conn, req *http.Request) *responseWriter {
	return &responseWriter{
		conn: conn, header: make(http.Header), req: req, body: bytes.NewBuffer(make([]byte, 0)),
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
func (r *responseWriter) finish() {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	if r.header.Get("Content-Length") == "" {
		r.header.Set("Content-Length", strconv.Itoa(r.body.Len()))
	}
	resp := bytes.NewBuffer(make([]byte, 0))
	is11 := r.req.ProtoAtLeast(1, 1)
	writeHeaderLine(resp, is11, r.status)
	for k, v := range r.header {
		for _, s := range v {
			_, _ = fmt.Fprintf(resp, "%s: %s\r\n", k, s)
		}
	}
	_, _ = fmt.Fprintf(resp, "\r\n")
	resp.Write(r.body.Bytes())
	r.body.Reset()
	_, _ = r.conn.Write(resp.Bytes())
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) error(err error) {
	if err == nil {
		return
	}
	r.req = &http.Request{}
	r.WriteHeader(http.StatusInternalServerError)
	_, _ = r.Write([]byte(err.Error()))
	r.finish()
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
