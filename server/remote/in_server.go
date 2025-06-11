package remote

import (
	"common/configs"
	"common/log"
	"common/remote"
	"encoding/json"
	"os"
	defin "server/define"
)

const isTunnelConnKey = "Tunnel-Conn"

const tunnelPort = "Tunnel-Port"

type InServer struct {
	remote.BaseServerHandler

	//Current server.
	server *remote.Server
}

func New() *InServer {
	return &InServer{
		server: nil,
	}
}

func (t *InServer) Reader(conn *remote.ConnV2, traverse remote.TraverseBy) {
	tunnelConn := t.isTunnelConn(conn)
	if !tunnelConn {
		//Determining whether a communication channel is connected.
		req, err := remote.Decoder(conn.GetReader())
		if err != nil {
			return
		}
		//Start process.
		inProcess(&req, conn)
	} else {
		port := t.getTunnelPort(conn)
		//Tunnel protocol
		if tunnel := defin.GetTunnel(port); tunnel != nil {
			tunnel.Receiver(conn)
		} else {
			log.Warn("Not found tunnel %d", port)
		}
	}
	traverse()
}
func (t *InServer) isTunnelConn(conn *remote.ConnV2) bool {
	attr, b := conn.GetContext().GetAttr(isTunnelConnKey)
	if b {
		return attr.(bool)
	}
	return false
}

func (t *InServer) getTunnelPort(conn *remote.ConnV2) int32 {
	attr, b := conn.GetContext().GetAttr(tunnelPort)
	if b {
		return attr.(int32)
	}
	return 0
}

func inProcess(p *remote.Protocol, conn *remote.ConnV2) {
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
	response := remote.NewResponse(p.Cmd, p.ReqId)
	data, err := entry.process(req, conn)
	if data != nil {
		byts, err := json.Marshal(data)
		response.Data = byts
		if err != nil {
			response.RspCode = remote.Rsp_fail
		}
	}
	if err != nil {
		response.RspCode = remote.Rsp_fail
	}
	outBytes := remote.Encoder(response)
	newConn := transform(conn, req)
	_, err = newConn.Write(outBytes)
	if err != nil {
		log.Warn("Writer %s , marshal json, error %s ", cmd, err.Error())
		return
	}
}

func transform(conn *remote.ConnV2,
	req remote.InBound) *remote.ConnV2 {
	if req.Cmd() == remote.Register {
		switch req.(type) {
		case remote.RegisterReq:
			bindId := req.(remote.RegisterReq).BindId
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
	t.server = remote.NewServer(cf.ServerPort)
	t.server.AddHandler(
		//remote.NewIdleServerHandler(5*time.Second),
		t,
	)
	err := t.server.Start(remote.WithSmun(remote.DefaulServerSmux()))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

type handlerEntry struct {
	newRequest func(data []byte) (remote.InBound, error)
	process    func(request remote.InBound, conn *remote.ConnV2) (any, error)
}

var handlers = make(map[remote.Cmd]handlerEntry)

// Register  [T inter.InBound]
//
//	@Description: register process.
//	@param cmd
//	@param process
func Register[T remote.InBound](cmd remote.Cmd, process InProcess[T]) {
	handlers[cmd] = handlerEntry{
		newRequest: func(data []byte) (remote.InBound, error) {
			var req T
			err := json.Unmarshal(data, &req)
			return req, err
		},
		process: func(r remote.InBound, conn *remote.ConnV2) (any, error) {
			req := r.(T)
			return process(req, conn)
		},
	}
}
