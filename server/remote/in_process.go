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

package remote

import (
	"fmt"
	"time"

	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/server/defin"
	"github.com/brook/server/tunnel"
)

func init() {
	Register(exchange.Heart, pingProcess)
	Register(exchange.Register, registerProcess)
	Register(exchange.LoginTunnel, loginProcess)
	Register(exchange.OpenTunnel, openTunnelProcess)
	Register(exchange.UdpRegister, dupRegisterProcess)
}

type InProcess[T exchange.InBound] func(request T, ch transport.Channel) (any, error)

// pingProcess handles ping/pong heartbeat messages between servers
// It takes a heartbeat request and a transport channel as input
// and returns a response heartbeat or an error
func pingProcess(request exchange.Heartbeat, ch transport.Channel) (any, error) {
	// Log the received ping message with its value and remote address
	log.Debug("Receiver Ping message : %s:%v", request.Value, ch.RemoteAddr())
	// Create a heartbeat response with PONG value
	// preserving the original start time and adding current server time
	heartbeat := exchange.Heartbeat{Value: "PONG",
		StartTime:  request.StartTime,
		ServerTime: time.Now().UnixMilli(),
	}
	return heartbeat, nil
}

func dupRegisterProcess(request exchange.UdpRegisterReqAndRsp, ch transport.Channel) (any, error) {
	return doRegister(request, ch)
}

// registerProcess handles the registration of a tunnel connection
// It takes a request containing registration details and a transport channel
// Returns the processed request and any error that occurred during registration
func registerProcess(request exchange.RegisterReqAndRsp, ch transport.Channel) (any, error) {
	return doRegister(request, ch)
}

func doRegister(request exchange.TRegister, ch transport.Channel) (any, error) {
	// Check the type of the channel and perform channel-specific operations
	switch sch := ch.(type) {
	case *transport.SChannel:
		// If it's a secure channel, mark it as a tunnel and add the proxy ID attribute
		sch.IsOpenTunnel = true
		sch.AddAttr(defin.HttpIdKey, request.GetHttpId())
		sch.AddAttr(defin.ProxyIdKey, request.GetProxyId())
	default:
		// Log error and return error for unsupported channel types
		log.Error("Not support channel type: %T", ch)
		return nil, fmt.Errorf("not support channel type:%T", ch)
	}
	// Bind the channel ID to the request
	//request.BindId = ch.GetId()
	// Get the tunnel using the specified port
	port := request.GetTunnelPort()
	t := tunnel.GetTunnel(port)
	if t == nil {
		// Log error and return error if tunnel is not found
		log.Error("Not found tunnel: %d", port)
		return nil, fmt.Errorf("not found tunnel:%d", port)
	}
	// Log debug information about the tunnel being registered
	log.Debug("Registering tunnel:%v", t)
	// Register the connection with the tunnel
	t.RegisterConn(ch, request)
	// Return the processed request
	return request, nil
}

func loginProcess(req exchange.LoginReq, ch transport.Channel) (any, error) {
	token := defin.GetToken()
	if token != req.Token {
		log.Warn("token not match,1:%v,2:%v", token, req.Token)
		return nil, fmt.Errorf("token not match")
	}
	port := defin.Get[int](defin.TunnelPortKey)
	return exchange.LoginResp{
		TunnelPort: port,
		UnId:       ch.GetId(),
	}, nil
}

func openTunnelProcess(req exchange.OpenTunnelReq, ch transport.Channel) (any, error) {
	cfg, err := OpenTunnelServer(req, ch)
	if err != nil {
		return nil, err
	}
	return exchange.OpenTunnelResp{
		UnId:        req.UnId,
		RemotePort:  cfg.RemotePort,
		Destination: cfg.Destination,
	}, nil
}
