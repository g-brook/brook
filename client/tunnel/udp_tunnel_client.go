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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/iox"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	"github.com/brook/common/transport"
)

type UdpTunnelClient struct {
	*clis.BaseTunnelClient
	reconnect    *clis.ReconnectManager
	multipleTcp  *MultipleTunnelClient
	localAddress *net.UDPAddr
	bufSize      int
	udpConnMap   *hash.SyncMap[string, *net.UDPConn]
}

func NewUdpTunnelClient(cfg *configs.ClientTunnelConfig, mtpc *MultipleTunnelClient) *UdpTunnelClient {
	if cfg.UdpSize == 0 {
		cfg.UdpSize = 1500
	}
	tunnelClient := clis.NewBaseTunnelClient(cfg, false)
	client := UdpTunnelClient{
		BaseTunnelClient: tunnelClient,
		multipleTcp:      mtpc,
		bufSize:          cfg.UdpSize,
		udpConnMap:       hash.NewSyncMap[string, *net.UDPConn](),
	}
	var err error
	client.localAddress, err = net.ResolveUDPAddr("udp", cfg.Destination)
	if err != nil {
		log.Error("NewUdpTunnelClient error %v", err)
		return nil
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	client.reconnect = clis.NewReconnectionManager(3 * time.Second)
	return &client
}

func (t *UdpTunnelClient) GetName() string {
	return "udp"
}

func (t *UdpTunnelClient) initOpen(*transport.SChannel) (err error) {

	stop := make(chan struct{})
	readLoop := func(updConn *net.UDPConn, remoteAddress *net.UDPAddr, bucket *exchange.TunnelBucket) {
		pool := iox.GetByteBufPool(t.bufSize)
		for {
			err := iox.WithBuffer(func(buf []byte) error {
				_, _, err = updConn.ReadFromUDP(buf)
				if err != nil {
					return err
				}
				pk := exchange.NewUdpPackage(buf, nil, remoteAddress)
				data, err := json.Marshal(pk)
				_ = bucket.Push(data, nil)
				return err
			}, pool)
			if err != nil && err == io.EOF {
				close(stop)
				return
			}
			select {
			case <-bucket.Done():
				close(stop)
				return
			default:
			}
		}
	}

	revLoop := func(rw io.ReadWriteCloser, bucket *exchange.TunnelBucket) {
		bucket.DefaultRead(func(p *exchange.TunnelProtocol) {
			data := p.Data
			var pk exchange.UdpPackage
			err = json.Unmarshal(data, &pk)
			if err != nil {
				close(stop)
				return
			}
			connKey := pk.RemoteAddress.String()
			udpConn, ok, err := t.localConn(connKey)
			if err != nil {
				if udpConn != nil {
					_ = udpConn.Close()
				}
				log.Error("%v", err)
				close(stop)
				return
			}
			_, err2 := udpConn.Write(p.Data)
			if err2 != nil {
				log.Error("Write to local address error %v", err2)
				close(stop)
				return
			}
			if !ok {
				threading.GoSafe(func() {
					readLoop(udpConn, pk.RemoteAddress, bucket)
				})
			}
		})
	}
	err = t.AsyncRegister(t.getReq(), func(p *exchange.Protocol, rw io.ReadWriteCloser, _ context.Context) error {
		if p.IsSuccess() {
			log.Info("Connection local address success then Client to server register success:%v", t.GetCfg().Destination)
			bucket := exchange.NewTunnelBucket(rw, t.TcControl.Context())
			revLoop(rw, bucket)
			bucket.Run()
			<-stop
		} else {
			log.Error("Connection local address success then Client to server register fail:%v", t.GetCfg().Destination)
			return fmt.Errorf("register fail")
		}
		return nil
	})
	if err != nil {
		log.Error("Connection fail %v", err)
		return err
	}
	return nil
}
func (t *UdpTunnelClient) getReq() *exchange.UdpRegisterReqAndRsp {
	return &exchange.UdpRegisterReqAndRsp{
		RegisterReqAndRsp: t.GetRegisterReq(),
		RemoteAddress:     t.localAddress.String(),
	}
}
func (t *UdpTunnelClient) localConn(connKey string) (*net.UDPConn, bool, error) {
	load, b := t.udpConnMap.Load(connKey)
	if b {
		if t.isConnAlive(load) {
			return load, true, nil
		}
		t.udpConnMap.Delete(connKey)
	}
	connFunction := func() (*net.UDPConn, error) {
		dial, err := net.DialUDP(string(lang.NetworkUdp), nil, t.localAddress)
		if err != nil {
			return nil, err
		}
		log.Info("Connection localAddress, %v success", t.GetCfg().Destination)
		t.udpConnMap.Store(connKey, dial)
		return dial, nil
	}
	dial, err := connFunction()
	return dial, false, err
}

// 检测 UDP 连接是否可用
func (t *UdpTunnelClient) isConnAlive(conn *net.UDPConn) bool {
	if conn == nil {
		return false
	}
	err := conn.SetWriteDeadline(time.Now().Add(1 * time.Millisecond))
	if err != nil {
		return false
	}
	_ = conn.SetWriteDeadline(time.Time{})
	return true
}
