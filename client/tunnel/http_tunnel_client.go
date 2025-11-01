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
	"context"
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
	"github.com/gobwas/ws"
)

func NewHttpTunnelClient(config *configs.ClientTunnelConfig) *HttpTunnelClient {
	if config.HttpId == "" {
		panic("httpId is empty")
	}
	tunnelClient := clis.NewBaseTunnelClient(config, true)
	client := HttpTunnelClient{
		BaseTunnelClient: tunnelClient,
	}
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
func (h *HttpTunnelClient) bindHandler(_ *exchange.Protocol, rw io.ReadWriteCloser, ctx context.Context) error {
	// closeConn is a helper function to close network connections
	closeConn := func(conn net.Conn) {
		if conn != nil {
			_ = conn.Close()
		}
	}
	loopRead := func() error {
		pt := exchange.NewTunnelRead()
		err := pt.Read(rw)
		if err != nil {
			return err
		}
		// If the pt.Ver is v1, that it is a http request
		if pt.Ver == exchange.V1 || pt.Ver == exchange.WebsocketV1 {
			buf := new(bytes.Buffer)
			buf.Write(pt.Data)
			if !isHTTPRequestCompleteLight(buf) {
				return nil
			}
			response, dial, err := httpCall(h.GetCfg().LocalAddress, pt.Data)
			fmt.Println("接收成功，建立连接........")
			//to websocket ss.
			if pt.Ver == exchange.WebsocketV1 {
				if err != nil {
					CloseWebsocketLeft(rw, pt.ReqId)
					return nil
				}
				path := string(pt.Attr)
				NewWebsocketClient(ctx, path, dial, rw, pt.ReqId).toAdd()
				log.Debug("Connect to %v websocket success.")
				return nil
			} else {
				//to http ss.
				if err != nil {
					return err
				}
				defer closeConn(dial)
				defer buf.Reset()
				if response != nil {
					bodyBytes, _ := io.ReadAll(response.Body)
					response.Body.Close()
					response.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
					headerBytes := BuildCustomHTTPHeader(response)
					merged := append(headerBytes, bodyBytes...)
					return exchange.NewTunnelWriter(merged, pt.ReqId).Writer(rw)
				} else {
					log.Warn("Read request fail", err)
					return err
				}
			}
		} else if pt.Ver == exchange.WebsocketV2 {
			websocketCall(pt, rw)
		}
		return nil

	}
	// Main loop to handle incoming requests
	for {
		select {
		// Check for context cancellation
		case <-ctx.Done():
			return nil
		default:
		}
		// Process next request
		err := loopRead()
		if err == io.EOF {
			log.Debug("http stream close.")
			rw.Close()
		}
	}

}

func websocketCall(protocol *exchange.TunnelProtocol, left io.ReadWriteCloser) {
	if len(protocol.Attr) <= 1 {
		return
	}
	opCode := ws.OpCode(protocol.Attr[0])
	path := string(protocol.Attr[1:])
	conn, b := GetWebsocketClient(path)
	if !b {
		CloseWebsocketLeft(left, protocol.ReqId)
		return
	}
	conn.WriteToRight(opCode, protocol.Data)
}

func httpCall(address string, request []byte) (rsp *http.Response, dial net.Conn, err error) {
	dial, err = net.Dial("tcp", address)
	if err != nil {
		rsp = getErrorResponse()
		return
	}
	_, err = dial.Write(request)
	if err != nil {
		fmt.Println(err.Error())
		rsp = getErrorResponse()
		return
	}
	rsp, err = http.ReadResponse(bufio.NewReader(dial), nil)
	if err != nil {
		rsp = getErrorResponse()
		return
	}
	return
}

// isHTTPRequestCompleteLight checks if an HTTP request is complete by examining the buffer
// It determines completeness by checking if headers are present and if the body length matches Content-Length
// This is a lightweight version that doesn't fully parse the request
func isHTTPRequestCompleteLight(buf *bytes.Buffer) bool {
	data := buf.Bytes()
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	if idx == -1 {
		return false
	}
	headerPart := data[:idx+4]
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(headerPart)))
	if err != nil {
		return false
	}
	if req.ContentLength > 0 {
		return int64(len(data)-(idx+4)) >= req.ContentLength
	}
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
