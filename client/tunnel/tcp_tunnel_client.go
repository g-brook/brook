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
	"fmt"
	"io"
	"net"

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/iox"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
)

type TcpTunnelClient struct {
	*clis.BaseTunnelClient
}

func NewTcpTunnelClient(config *configs.ClientTunnelConfig, _ *MultipleTunnelClient) (*TcpTunnelClient, error) {
	tunnelClient := clis.NewBaseTunnelClient(config, false)
	client := TcpTunnelClient{
		BaseTunnelClient: tunnelClient,
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	return &client, nil
}

func (t *TcpTunnelClient) GetName() string {
	return "TcpTunnelClient"
}

func (t *TcpTunnelClient) initOpen(ch *transport.SChannel) error {
	localConnection, err := t.localConnection()
	if err != nil {
		if localConnection != nil {
			_ = localConnection.Close()
		}
		_ = ch.Close()
		return err
	}
	err = t.AsyncRegister(t.GetRegisterReq(), func(p *exchange.Protocol, rw io.ReadWriteCloser, _ context.Context) error {
		log.Info("Connection local address success then Client to server register success:%v", t.GetCfg().Destination)
		if p.IsSuccess() {
			addHealthyCheckStream(ch)
			errors := iox.Pipe(ch, localConnection)
			if len(errors) > 0 {
				log.Error("Pipe error %v", errors)
			}
			return nil
		} else {
			log.Error("Connection local address success then Client to server register fail:%v", t.GetCfg().Destination)
			return fmt.Errorf("register fail")
		}
	})
	if err != nil {
		if localConnection != nil {
			_ = localConnection.Close()
		}
		_ = ch.Close()
		log.Error("Connection fail %v", err)
		return err
	}
	return nil
}
func (t *TcpTunnelClient) localConnection() (net.Conn, error) {
	connFunction := func() (net.Conn, error) {
		dial, err := net.Dial(string(lang.NetworkTcp), t.GetCfg().Destination)
		if err != nil {
			return nil, err
		}
		log.Info("Connection localAddress, %v success", t.GetCfg().Destination)
		return dial, err
	}
	return connFunction()
}
