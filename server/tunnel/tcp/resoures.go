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

package tcp

import (
	"errors"

	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/log"
	trp "github.com/g-brook/brook/common/transport"
	. "github.com/g-brook/brook/server/tunnel"
)

type Resources struct {
	pool       *TunnelPool
	cfg        *configs.ServerTunnelConfig
	getManager func() trp.Channel
}

// NewResources creates and returns a new Resources instance
// This is a constructor function that initializes a Resources struct
func NewResources(size int,
	cfg *configs.ServerTunnelConfig,
	getManager func() trp.Channel) *Resources {
	p := &Resources{
		getManager: getManager,
		cfg:        cfg,
	}
	p.pool = NewTunnelPool(p.createConnection, size)
	return p
}

func (htl *Resources) createConnection() error {
	manager := htl.getManager()
	if manager != nil {
		req := &exchange.WorkConnReq{
			ProxyId: htl.cfg.Id,
		}
		log.Debug("create work conn request %v to %v", req, manager.RemoteAddr())
		request, _ := exchange.NewRequest(req)
		_, err := manager.Write(request.Bytes())
		if err != nil {
			log.Error("write work conn request %v to %v err %v", request, manager.RemoteAddr(), err)
		}
		return err
	}
	return errors.New("manager is nil, can't create connection")
}

func (htl *Resources) get() (trp.Channel, error) {
	return htl.pool.Get()
}

func (htl *Resources) put(ch trp.Channel) error {
	return htl.pool.Put(ch)
}
