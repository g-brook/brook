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

package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/httpx"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
)

func NewHttpTunnelClient(config *configs.ClientTunnelConfig) *HttpTunnelClient {
	if config.HttpId == "" {
		panic("httpId is empty")
	}
	tunnelClient := clis.NewBaseTunnelClient(config, true)
	// Create a new HttpTunnelClient embedding the base tunnel client
	client := HttpTunnelClient{
		BaseTunnelClient: tunnelClient,
	}
	// Assign the initOpen method to the DoOpen function pointer of the base tunnel client
	// This method will be called when the tunnel is opened
	tunnelClient.DoOpen = client.initOpen
	return &client
}

// HttpTunnelClient is a tunnel client that handles HTTP connections.
type HttpTunnelClient struct {
	*clis.BaseTunnelClient
}

// GetName returns the name of the tunnel client.
func (h *HttpTunnelClient) GetName() string {
	return "HttpTunnelClient"
}

// initOpen initializes the HTTP tunnel client by registering it and logging the result.
// Parameters:
//   - stream: The smux stream to use.
//
// Returns:
//   - error: An error if the registration fails.
func (h *HttpTunnelClient) initOpen(*transport.SChannel) error {
	h.BaseTunnelClient.AddReadHandler(exchange.WorkerConnReq, h.bindHandler)
	rsp, err := h.Register(h.GetRegisterReq())
	if err != nil {
		log.Error("Register fail %v", err)
		return err
	} else {
		log.Info("Register success:PORT-%v", rsp.TunnelPort)
	}
	return nil
}

// bindHandler handles the binding of HTTP tunnel client requests
func (h *HttpTunnelClient) bindHandler(_ *exchange.Protocol, rw io.ReadWriteCloser) error {
	// closeConn is a helper function to close network connections
	closeConn := func(conn net.Conn) {
		if conn != nil {
			_ = conn.Close()
		}
	}
	// call processes the HTTP request and returns the response and connection
	call := func(request []byte, err error) (rsp *http.Response, dial net.Conn) {
		if err != nil {
			return
		}
		// Establish TCP connection to local address
		dial, err = net.Dial("tcp", h.GetCfg().LocalAddress)
		if err != nil {
			rsp = getErrorResponse()
			return
		}
		// Write request to the local connection
		_, err = dial.Write(request)
		if err != nil {
			fmt.Println(err.Error())
			rsp = getErrorResponse()
			return
		}
		// Read HTTP response from the local connection
		rsp, err = http.ReadResponse(bufio.NewReader(dial), nil)
		if err != nil {
			rsp = getErrorResponse()
			return
		}
		return
	}
	// Create a new buffer to store request data
	buf := new(bytes.Buffer)
	// loopRead continuously reads and processes incoming requests
	loopRead := func() error {
		// Create a new tunnel reader for each request
		pt := exchange.NewTunnelRead()
		err := pt.Read(rw)
		if err != nil {
			return err
		}
		// Append received data to buffer
		buf.Write(pt.Data)
		// Check if HTTP request is complete
		if !isHTTPRequestCompleteLight(buf) {
			return nil
		}
		// Process the complete request
		response, dial := call(buf.Bytes(), err)
		defer closeConn(dial)
		defer buf.Reset()
		if response != nil {
			// Read response body
			bodyBytes, _ := io.ReadAll(response.Body)
			response.Body.Close()
			// Update content length header
			response.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
			// Build custom HTTP headers
			headerBytes := BuildCustomHTTPHeader(response)
			// Merge headers and body
			merged := append(headerBytes, bodyBytes...)
			// Write response back through tunnel
			return exchange.NewTunnelWriter(merged, pt.ReqId).Writer(rw)
		} else {
			log.Warn("Read request fail", err)
			return err
		}
	}
	// Main loop to handle incoming requests
	for {
		select {
		// Check for context cancellation
		case <-h.TcControl.Context().Done():
			return nil
		default:
		}
		// Process next request
		err := loopRead()
		if err == io.EOF {
			h.Close()
		}
	}

}

// isHTTPRequestCompleteLight checks if an HTTP request is complete by examining the buffer
// It determines completeness by checking if headers are present and if the body length matches Content-Length
// This is a lightweight version that doesn't fully parse the request
func isHTTPRequestCompleteLight(buf *bytes.Buffer) bool {
	// Read all data from the buffer
	data := buf.Bytes()
	// Find the end of headers marker (\r\n\r\n)
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	if idx == -1 {
		return false
	}
	// Extract the header part of the request
	headerPart := data[:idx+4]
	// Try to parse the headers from the buffer
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(headerPart)))
	if err != nil {
		return false
	}

	// If there's a Content-Length specified, check if we've received all the body data
	if req.ContentLength > 0 {
		return int64(len(data)-(idx+4)) >= req.ContentLength
	}
	// If no Content-Length is specified, assume the request is complete after headers
	return true
}

// BuildCustomHTTPHeader constructs a custom HTTP header from an HTTP response
// It formats the headers into a byte slice following the HTTP protocol standard
// Parameters:
//
//	r - pointer to the httpx.Response object containing the response data
//
// Returns:
//
//	[]byte - formatted HTTP header as a byte slice
func BuildCustomHTTPHeader(r *http.Response) []byte {
	// Create a buffer to build the HTTP header
	var buf bytes.Buffer

	// Format and write the status line (e.g., HTTP/1.1 200 OK)
	// Includes protocol version, status code, and status text
	st := fmt.Sprintf("HTTP/%d.%d %03d %s\r\n", r.ProtoMajor, r.ProtoMinor, r.StatusCode, r.Status)
	buf.WriteString(st)
	for key, value := range r.Header {
		if key != "Transfer-Encoding" {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value[0]))
		}
	}
	buf.WriteString("\r\n")
	return buf.Bytes()
}

func getErrorResponse() *http.Response {
	return httpx.GetResponse(http.StatusInternalServerError)

}
