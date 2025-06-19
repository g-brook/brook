package remote

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	defin "github.com/brook/server/define"
	"os"
)

const isTunnelConnKey = "Tunnel-Conn"

const tunnelPort = "Tunnel-Port"

type InServer struct {
	srv.BaseServerHandler

	//Current server.
	server *srv.Server
}

func New() *InServer {
	return &InServer{
		server: nil,
	}
}

func (t *InServer) Reader(ch srv.Channel, traverse srv.TraverseBy) {
	switch ch.(type) {
	case *srv.GChannel:
		//Determining whether a communication channel is connected.
		req, err := exchange.Decoder(ch.GetReader())
		if err != nil {
			return
		}
		//Start process.
		inProcess(req, ch.(*srv.GChannel))
	case *srv.SChannel:
		fmt.Println("In Server")
	}

	traverse()
}
func (t *InServer) isTunnelConn(conn *srv.GChannel) bool {
	attr, b := conn.GetContext().GetAttr(isTunnelConnKey)
	if b {
		return attr.(bool)
	}
	return false
}

func (t *InServer) getTunnelPort(conn *srv.GChannel) int32 {
	attr, b := conn.GetContext().GetAttr(tunnelPort)
	if b {
		return attr.(int32)
	}
	return 0
}

func inProcess(p *exchange.Protocol, conn *srv.GChannel) {
	cmd := p.Cmd
	entry, ok := handlers[cmd]
	if !ok {
		log.Warn("Unknown cmd %s ", cmd)
		return
	}
	req, err := entry.newRequest(p.Data)
	if err != nil {
		log.Warn("Cmd %s , unmarshal json, error %s ", cmd, err.Error())
		return
	}
	response, _ := exchange.NewResponse(p.Cmd, p.ReqId)
	data, err := entry.process(req, conn)
	if data != nil {
		byts, err := json.Marshal(data)
		response.Data = byts
		if err != nil {
			response.RspCode = exchange.RspFail
		}
	}
	if err != nil {
		response.RspCode = exchange.RspFail
	}
	outBytes := exchange.Encoder(response)
	newConn := transform(conn, req)
	writer := bufio.NewWriterSize(newConn, len(outBytes))
	_, err = writer.Write(outBytes)
	if err != nil {
		log.Warn("Writer %s , marshal json, error %s ", cmd, err.Error())
		return
	}
_:
	writer.Flush()
}

func transform(conn *srv.GChannel,
	req exchange.InBound) *srv.GChannel {
	if req.Cmd() == exchange.Register {
		switch req.(type) {
		case exchange.RegisterReq:
			bindId := req.(exchange.RegisterReq).BindId
			newConn, ok := conn.GetServer().GetConnection(bindId)
			if ok {
				return newConn
			}
		}
	}
	return conn
}

// Start
//
//	@Description:  Start In Server. Port is between 4000 and  9000.
//	@receiver t
//	@param cf config.
func (t *InServer) Start(cf *configs.ServerConfig) *InServer {
	//Judgment server port lt 4000 or gt 9000,otherwise setting serve port 7000
	if cf.ServerPort < 4000 || cf.ServerPort > 9000 {
		cf.ServerPort = configs.DefServerPort
	}
	//Start local server.
	go t.onStart(cf)
	log.Info("Start the In-Server ,port is : %d ", cf.ServerPort)
	return t
}

func (t *InServer) onStart(cf *configs.ServerConfig) {
	go t.onStartServer(cf)
	go t.onStartTunnelServer(cf)
}

func (t *InServer) onStartServer(cf *configs.ServerConfig) {
	t.server = srv.NewServer(cf.ServerPort)
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
	t.server = srv.NewServer(port)
	t.server.AddHandler(t)
	defin.Set(defin.TunnelPortKey, port)
	err := t.server.Start(srv.WithServerSmux(srv.DefaultServerSmux()))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

type handlerEntry struct {
	newRequest func(data []byte) (exchange.InBound, error)
	process    func(request exchange.InBound, conn *srv.GChannel) (any, error)
}

var handlers = make(map[exchange.Cmd]handlerEntry)

// Register  [T inter.InBound]
//
//	@Description: register process.
//	@param cmd
//	@param process
func Register[T exchange.InBound](cmd exchange.Cmd, process InProcess[T]) {
	handlers[cmd] = handlerEntry{
		newRequest: func(data []byte) (exchange.InBound, error) {
			var req T
			err := json.Unmarshal(data, &req)
			return req, err
		},
		process: func(r exchange.InBound, conn *srv.GChannel) (any, error) {
			req := r.(T)
			return process(req, conn)
		},
	}
}
