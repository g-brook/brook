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
	"context"
	"encoding/json"
	"os"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	"github.com/brook/common/transport"
	"github.com/brook/server/defin"
	srv2 "github.com/brook/server/srv"
)

const isTunnelConnKey = "TunnelServer-conn"

const tunnelPort = "TunnelServer-Port"

type InServer struct {
	srv2.BaseServerHandler

	//Current server.
	server *srv2.Server

	//tunnelServer
	tunnelServer *srv2.Server
}

func New() *InServer {
	return &InServer{
		server: nil,
	}
}

func (t *InServer) Reader(ch transport.Channel, traverse srv2.TraverseBy) {
	//Determining whether a communication channel is connected.
	req, err := exchange.Decoder(ch.GetReader())
	if err != nil {
		return
	}
	//Start process.
	inProcess(req, ch)
	traverse()
}

func (t *InServer) isTunnelConn(conn *srv2.GChannel) bool {
	attr, b := conn.GetContext().GetAttr(isTunnelConnKey)
	if b {
		return attr.(bool)
	}
	return false
}

func (t *InServer) getTunnelPort(conn *srv2.GChannel) int32 {
	attr, b := conn.GetContext().GetAttr(tunnelPort)
	if b {
		return attr.(int32)
	}
	return 0
}

// Shutdown This function shuts down the InServer instance
func (t *InServer) Shutdown() {
	// Check if the server is not nil
	if t.server != nil {
		// Shutdown the server
		t.server.Shutdown(context.Background())
	}
	// Check if the tunnel server is not nil
	if t.tunnelServer != nil {
		// Shutdown the tunnel server
		t.tunnelServer.Shutdown(context.Background())
	}
}

// GetConnection This function returns a connection from the server given an id
func (t *InServer) GetConnection(id string) (*srv2.GChannel, bool) {
	// Call the GetConnection function from the server with the given id
	return t.server.GetConnection(id)
}

// This function handles the processing of a request from a client
func inProcess(p *exchange.Protocol, conn transport.Channel) {
	// Get the command from the protocol
	cmd := p.Cmd
	// Check if the command is known
	entry, ok := handlers[cmd]
	if !ok {
		// If the command is not known, log a warning and return
		log.Warn("Unknown cmd %s ", cmd)
		return
	}
	// Create a new request from the data in the protocol
	req, err := entry.newRequest(p.Data)
	if err != nil {
		// If there is an error creating the request, log a warning and return
		log.Warn("Cmd %s , unmarshal json, error %s ", cmd, err.Error())
		return
	}
	// Create a new response with the command and request ID from the protocol
	response, _ := exchange.NewResponse(p.Cmd, p.ReqId)
	// Process the request and get the response data
	data, err := entry.process(req, conn)
	if !entry.isResponse() {
		return
	}
	if data != nil {
		// If there is data in the response, marshal it into bytes
		byts, err := json.Marshal(data)
		response.Data = byts
		// If there is an error marshaling the data, set the response code to fail
		if err != nil {
			response.RspCode = exchange.RspFail
		}
	}
	// If there is an error processing the request, set the response code to fail
	if err != nil {
		response.RspCode = exchange.RspFail
		response.RspMsg = err.Error()
	}
	// Encode the response into bytes
	outBytes := exchange.Encoder(response)
	// Write the encoded response to the connection
	_, err = conn.Write(outBytes)
	// If there is an error writing the response, log a warning
	if err != nil {
		log.Warn("Writer %s , marshal json, error %s ", cmd, err.Error())
		return
	}
}

// Start
//
//	@Description:  Start In Server. Port is between 4000 and  9000.
//	@receiver t
//	@param cf configs.
func (t *InServer) Start(cf *configs.ServerConfig) *InServer {
	//Judgment server port lt 4000 or gt 9000,otherwise setting serve port 7000
	if cf.ServerPort < 4000 || cf.ServerPort > 9000 {
		cf.ServerPort = configs.DefServerPort
	}
	//Start local server.
	threading.GoSafe(func() {
		t.onStart(cf)
	})
	log.Info("Start the In-Server ,port is : %d ", cf.ServerPort)
	return t
}

// This function starts the InServer by starting the onStartServer and onStartTunnelServer functions in separate goroutines
func (t *InServer) onStart(cf *configs.ServerConfig) {
	// Start the onStartServer function in a separate goroutine
	threading.GoSafe(func() {
		t.onStartServer(cf)
	})
	// Start the onStartTunnelServer function in a separate goroutine
	threading.GoSafe(func() {
		t.onStartTunnelServer(cf)
	})
}

func (t *InServer) onStartServer(cf *configs.ServerConfig) {
	t.server = srv2.NewServer(cf.ServerPort)
	t.server.AddHandler(t)
	err := t.server.Start()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func (t *InServer) onStartTunnelServer(cf *configs.ServerConfig) {
	port := cf.TunnelPort
	if port < cf.ServerPort {
		port = cf.ServerPort + 10
		cf.TunnelPort = port
	}
	t.tunnelServer = srv2.NewServer(port)
	t.tunnelServer.AddHandler(t)
	defin.Set(defin.TunnelPortKey, port)
	err := t.tunnelServer.Start(srv2.WithServerSmux(srv2.DefaultServerSmux()))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

type handlerEntry struct {
	newRequest func(data []byte) (exchange.InBound, error)
	process    func(request exchange.InBound, conn transport.Channel) (any, error)
	isResponse func() bool
}

var handlers = make(map[exchange.Cmd]handlerEntry)

// Register  [T inter.InBound]
//
//	@Description: register process.
//	@param cmd
//	@param process
func Register[T exchange.InBound](cmd exchange.Cmd, process InProcess[T], isResponse bool) {
	handlers[cmd] = handlerEntry{
		newRequest: func(data []byte) (exchange.InBound, error) {
			var req T
			err := json.Unmarshal(data, &req)
			return req, err
		},
		process: func(r exchange.InBound, conn transport.Channel) (any, error) {
			req := r.(T)
			return process(req, conn)
		},
		isResponse: func() bool {
			return isResponse
		},
	}
}
