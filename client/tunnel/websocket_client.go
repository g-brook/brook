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

	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebsocketClientManager struct {
	clients *hash.SyncMap[string, *WebsocketBridge]
}

func newWebsocketClientManager() *WebsocketClientManager {
	return &WebsocketClientManager{
		clients: hash.NewSyncMap[string, *WebsocketBridge](),
	}
}

func (r *WebsocketClientManager) getWebsocketBridge(path string) (*WebsocketBridge, bool) {
	client, b := r.clients.Load(path)
	if b {
		return client, true
	}
	return nil, false
}

func (r *WebsocketClientManager) closeWebsocketLeft(left io.ReadWriteCloser, reqId int64) {
	if left == nil {
		return
	}
	attr := make([]byte, 1)
	attr[0] = byte(ws.OpClose)
	writer := exchange.NewTunnelWebsocketWriterV2([]byte{}, attr, reqId)
	_ = writer.Writer(left)
}

func (r *WebsocketClientManager) addWebsocketConn(client *WebsocketBridge) {
	if oldClient, b := r.clients.LoadOrStore(client.path, client); b {
		oldClient.Close(false)
	}
	r.clients.Store(client.path, client)
}

func (r *WebsocketClientManager) delWebsocketConn(path string) {
	r.clients.Delete(path)
}

func (r *WebsocketClientManager) newWebsocketBridge(
	ctx context.Context,
	path string,
	left io.ReadWriteCloser,
	right net.Conn,
	reqId int64) *WebsocketBridge {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &WebsocketBridge{
		path:      path,
		left:      left,
		right:     right,
		reqId:     reqId,
		cancelCtx: cancelCtx,
		cancel:    cancelFunc,
		manager:   r,
	}
}

type WebsocketBridge struct {
	path      string
	right     net.Conn
	left      io.ReadWriteCloser
	reqId     int64
	cancelCtx context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	manager   *WebsocketClientManager
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
		c.manager.delWebsocketConn(c.path)
		c.right = nil
		c.cancel()
		if isCloseLeft {
			c.manager.closeWebsocketLeft(c.left, c.reqId)
		}
	})
}

func (c *WebsocketBridge) toRunning() {
	c.manager.addWebsocketConn(c)
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
