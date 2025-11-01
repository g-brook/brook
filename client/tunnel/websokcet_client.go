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
	"context"
	"io"
	"net"
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var wssClients = hash.NewSyncMap[string, *WebsocketBridge]()

func GetWebsocketClient(path string) (*WebsocketBridge, bool) {
	client, b := wssClients.Load(path)
	if b {
		return client, true
	}
	return nil, false
}

func CloseWebsocketLeft(left io.ReadWriteCloser, reqId int64) {
	if left == nil {
		return
	}
	attr := make([]byte, 1)
	attr[0] = byte(ws.OpClose)
	writer := exchange.NewTunnelWebsocketWriterV2([]byte{}, attr, reqId)
	_ = writer.Writer(left)
}

func AddWebsocketConn(client *WebsocketBridge) {
	if oldClient, b := wssClients.LoadOrStore(client.path, client); b {
		oldClient.Close(false)
	}
	wssClients.Store(client.path, client)
}

func DelWebsocketConn(path string) {
	wssClients.Delete(path)
}

type WebsocketBridge struct {
	path      string
	right     net.Conn
	left      io.ReadWriteCloser
	reqId     int64
	cancelCtx context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func (c *WebsocketBridge) Write(p []byte) (n int, err error) {
	writer := exchange.NewTunnelWebsocketWriterV2(p, []byte{}, c.reqId)
	return len(p), writer.Writer(c.left)
}

func (c *WebsocketBridge) loop() {
	threading.GoSafe(func() {
		for {
			select {
			case <-c.cancelCtx.Done():
				c.Close(false)
			default:
			}
			if c.right == nil {
				return
			}
			data, _, err := wsutil.ReadServerData(c.right)
			if err != nil {
				log.Error("Right:Read from websocket fail:", err)
			}
			if err == io.EOF {
				c.Close(true)
				log.Debug("Right:Close from websocket")
				return
			}
			_, _ = c.Write(data)
		}
	})
}

func (c *WebsocketBridge) Close(isCloseLeft bool) {
	c.closeOnce.Do(func() {
		_ = c.right.Close()
		DelWebsocketConn(c.path)
		c.right = nil
		c.cancel()
		if isCloseLeft {
			CloseWebsocketLeft(c.left, c.reqId)
		}
	})
}

func (c *WebsocketBridge) toAdd() {
	AddWebsocketConn(c)
	//开启.
	c.loop()
}

func (c *WebsocketBridge) WriteToRight(code ws.OpCode, data []byte) {
	if c.right == nil {
		return
	}
	err := wsutil.WriteClientMessage(c.right, code, data)
	if err != nil {
		log.Error("Right:Write to websocket fail:", err)
		return
	}
	if code == ws.OpClose {
		c.Close(false)
	}
}

func NewWebsocketClient(
	ctx context.Context,
	path string,
	right net.Conn,
	left io.ReadWriteCloser,
	reqId int64) *WebsocketBridge {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &WebsocketBridge{
		path:      path,
		right:     right,
		left:      left,
		reqId:     reqId,
		cancelCtx: cancelCtx,
		cancel:    cancelFunc,
	}
}
