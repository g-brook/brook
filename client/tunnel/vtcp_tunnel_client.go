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

	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/transport"
)

type VTCPTunnelClient struct {
	*clis.BaseTunnelClient
}

func NewVTcpTunnelClient(config *configs.ClientTunnelConfig, _ *MultipleTunnelClient) (*VTCPTunnelClient, error) {
	tunnelClient := clis.NewBaseTunnelClient(config, true)
	client := VTCPTunnelClient{
		BaseTunnelClient: tunnelClient,
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	return &client, nil
}

func (c *VTCPTunnelClient) initOpen(ch *transport.SChannel) error {
	return c.AsyncVisitorRegister(nil, func(p *exchange.Protocol, rw io.ReadWriteCloser, ctx context.Context) error {
		if !p.IsSuccess() {
			return errors.New("vtcp tunnel open fail")
		}
		//open local server.
		return nil
	})
}
