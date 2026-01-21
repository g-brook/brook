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
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/httpx"
	"github.com/brook/common/ringbuffer"
	"github.com/brook/common/threading"
)

var httpError = errors.New("error: Agent connect to server failed")

type HttpClientManager struct {
	clients *hash.SyncMap[int64, *HttpBridge]
	lock    sync.Mutex
}

func NewHttpClientManager() *HttpClientManager {
	return &HttpClientManager{
		clients: hash.NewSyncMap[int64, *HttpBridge](),
	}
}

func (r *HttpClientManager) GetHttpBridge(ctx context.Context,
	left io.ReadWriteCloser,
	rightAddress string,
	reqId int64, isWs bool) (*HttpBridge, error) {
	load, b := r.clients.Load(reqId)
	if b {
		return load, nil
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	load, b = r.clients.Load(reqId)
	if b {
		return load, nil
	}
	dial, err := net.Dial("tcp", rightAddress)
	if err != nil {
		return nil, err
	}
	bridge := r.newHttpBridge(ctx, left, dial, reqId, r)
	bridge.isWs = isWs
	bridge.toRunning()
	r.clients.Store(reqId, bridge)
	return bridge, nil
}

func (r *HttpClientManager) newHttpBridge(ctx context.Context, left io.ReadWriteCloser, right net.Conn, reqId int64, manager *HttpClientManager) *HttpBridge {
	newCtx, cancel := context.WithCancel(ctx)
	return &HttpBridge{
		left:    left,
		right:   right,
		buffer:  ringbuffer.Get(),
		context: newCtx,
		cancel:  cancel,
		reqId:   reqId,
		manager: manager,
	}
}

type upgraderDo func()

type HttpBridge struct {
	left          io.ReadWriteCloser
	right         net.Conn
	buffer        *ringbuffer.RingBuffer
	context       context.Context
	cancel        context.CancelFunc
	reqId         int64
	manager       *HttpClientManager
	isRunning     atomic.Bool
	cancelOnce    sync.Once
	lastWriteTime time.Time
	isWs          bool
	hp            upgraderDo
}

func (b *HttpBridge) Read(p []byte) (n int, err error) {
	return b.buffer.Read(p)
}

func (b *HttpBridge) Write(p []byte) (n int, err error) {
	err = exchange.NewTunnelWriter(p, b.reqId).Writer(b.left)
	b.lastWriteTime = time.Now()
	return len(p), err
}

func (b *HttpBridge) toRunning() {
	//to writer right.
	if b.isRunning.Load() {
		return
	}
	defer b.isRunning.Store(true)
	threading.GoSafe(func() {
		select {
		case <-b.context.Done():
			b.Close()
			return
		default:
			_, err := io.Copy(b.right, b)
			if err == io.EOF {
				b.Close()
				return
			}
		}
	})
	threading.GoSafe(func() {
		select {
		case <-b.context.Done():
			b.Close()
			return
		default:
			defer b.Close()
			reader := bufio.NewReader(b.right)
			response, err := http.ReadResponse(reader, nil)
			if err != nil {
				errorResponse := getErrorResponse(httpError)
				_, err = b.Write(errorResponse)
				return
			}
			defer response.Body.Close()
			if b.isWs && b.hp != nil {
				if response.StatusCode == http.StatusSwitchingProtocols {
					b.hp()
					return
				}
			}
			bodyBytes, _ := io.ReadAll(response.Body)
			headerBytes := BuildCustomHTTPHeader(response, len(bodyBytes))
			merged := append(headerBytes, bodyBytes...)
			_, err = b.Write(merged)
		}
	})
}

func (b *HttpBridge) Close() {
	b.cancelOnce.Do(func() {
		if b.buffer != nil {
			b.buffer.Reset()
			ringbuffer.Put(b.buffer)
		}
		b.cancel()
		if !b.isWs {
			_ = b.right.Close()
		}
		b.manager.clients.Delete(b.reqId)
		b.manager = nil
		b.buffer = nil
		b.left = nil
		b.context = nil
		b.right = nil
	})
}

func (b *HttpBridge) WriterToRight(p []byte) (int, error) {
	return b.buffer.Write(p)
}

func (b *HttpBridge) websocket(ctx context.Context, path string, websocket *WebsocketClientManager) (*WebsocketBridge, error) {
	if !b.isWs {
		return nil, fmt.Errorf("not websocket protocol")
	}
	return websocket.newWebsocketBridge(ctx, path, b.left, b.right, b.reqId), nil
}

func (b *HttpBridge) upgrader(upd upgraderDo) {
	b.hp = upd
}

func getErrorResponse(err error) []byte {
	response := httpx.GetResponse(http.StatusInternalServerError)
	errMsg := []byte(err.Error())
	header := BuildCustomHTTPHeader(response, len(errMsg))
	result := make([]byte, 0, len(header)+len(errMsg))
	result = append(result, header...)
	result = append(result, errMsg...)
	return result
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
func BuildCustomHTTPHeader(r *http.Response, size int) []byte {
	var buf bytes.Buffer
	// Format and write the status line (e.g., HTTP/1.1 200 OK)
	// Includes protocol version, status code, and status text
	st := fmt.Sprintf("HTTP/%d.%d %03d %s\r\n", r.ProtoMajor, r.ProtoMinor, r.StatusCode, r.Status)
	buf.WriteString(st)
	r.Header.Del("Transfer-Encoding")
	r.Header.Set("Content-Length", strconv.Itoa(size))
	r.Header.Set("Connection", "close") // 非 keep-alive
	// 写入所有 header 字段
	for k, vv := range r.Header {
		for _, v := range vv {
			_, _ = fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
	}
	buf.WriteString("\r\n")
	return buf.Bytes()
}
