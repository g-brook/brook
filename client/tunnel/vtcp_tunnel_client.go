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
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/iox"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/transport"
	"github.com/xtaci/smux"
)

const visitorRegisterTimeout = 5 * time.Second

type VTCPTunnelClient struct {
	*clis.BaseTunnelClient
	closeOnce sync.Once
	localConn net.Conn
}

func (c *VTCPTunnelClient) OpenVisitor(session *smux.Session, localConn net.Conn) error {
	c.localConn = localConn
	c.DoOpen = c.openVisitorStream
	return c.Open(session)
}

func NewVTcpTunnelClient(config *configs.ClientTunnelConfig) (*VTCPTunnelClient, error) {
	tunnelClient := clis.NewBaseTunnelClient(config, false)
	client := VTCPTunnelClient{
		BaseTunnelClient: tunnelClient,
	}
	return &client, nil
}

func (c *VTCPTunnelClient) GetName() string {
	return "VTcpTunnelClient"
}

func (c *VTCPTunnelClient) openVisitorStream(stream *transport.SChannel) error {
	defer func() {
		_ = c.localConn.Close()
		_ = stream.Close()
	}()
	finish := make(chan error, 1)
	notifyFinish := func(err error) {
		select {
		case finish <- err:
		default:
		}
	}
	err := c.AsyncVisitorRegister(c.GetVisitorReq(), func(p *exchange.Protocol, nch io.ReadWriteCloser, _ context.Context) error {
		if !p.IsSuccess() {
			log.Error("VTCP visitor register fail:%s", p.RspMsg)
			notifyFinish(exchange.CloseError)
			return exchange.CloseError
		}
		addHealthyCheckStream(stream)
		errs := iox.Pipe(nch, c.localConn)
		if len(errs) > 0 {
			log.Debug("VTCP visitor pipe exit:%v", errs)
			notifyFinish(exchange.CloseError)
			return exchange.CloseError
		}
		notifyFinish(nil)
		return nil
	})
	if err != nil {
		return err
	}
	timer := time.NewTimer(visitorRegisterTimeout)
	defer timer.Stop()
	select {
	case err = <-finish:
		return err
	case <-timer.C:
		return errors.New("visitor register timeout")
	case <-stream.Done():
		return nil
	case <-c.Done():
		return nil
	}
}
