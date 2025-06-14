package remote

import (
	"fmt"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	defin "github.com/brook/server/define"
)

func init() {
	Register(srv.Heart, pingProcess)
	Register(srv.Register, registerProcess)
	Register(srv.Communication, communicationProcess)
}

type InProcess[T srv.InBound] func(request T, conn *srv.ConnV2) (any, error)

func pingProcess(request srv.Heartbeat, conn *srv.ConnV2) (any, error) {
	log.Info("Receiver Ping message : %s:%s", request.Value, conn.RemoteAddr().String())
	heartbeat := srv.Heartbeat{Value: "PONG"}
	return heartbeat, nil
}

func registerProcess(request srv.RegisterReq, conn *srv.ConnV2) (any, error) {
	port := request.TunnelPort
	tunnel := defin.GetTunnel(port)
	if tunnel == nil {
		log.Error("Not found tunnel: %d", port)
		return nil, fmt.Errorf("not found tunnel:%d", port)
	}
	tunnel.RegisterConn(conn, request)
	//Register conn to tunnel success.
	conn.GetContext().AddAttr(isTunnelConnKey, true)
	conn.GetContext().AddAttr(tunnelPort, port)
	return nil, nil
}

func communicationProcess(req srv.CommunicationInfo, conn *srv.ConnV2) (any, error) {
	id := conn.GetContext().Id
	req.BindId = id
	return req, nil
}
