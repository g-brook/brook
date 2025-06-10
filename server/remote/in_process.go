package remote

import (
	"common/log"
	"common/remote"
	"fmt"
	defin "server/define"
)

func init() {
	Register(remote.Heart, pingProcess)
	Register(remote.Register, registerProcess)
	Register(remote.Communication, communicationProcess)
}

type InProcess[T remote.InBound] func(request T, conn *remote.ConnV2) (any, error)

func pingProcess(request remote.Heartbeat, conn *remote.ConnV2) (any, error) {
	log.Debug("Receiver Ping message : %s:%s", request.Value, conn.RemoteAddr().String())
	heartbeat := remote.Heartbeat{Value: "PONG"}
	return heartbeat, nil
}

func registerProcess(request remote.RegisterReq, conn *remote.ConnV2) (any, error) {
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

func communicationProcess(req remote.CommunicationInfo, conn *remote.ConnV2) (any, error) {
	id := conn.GetContext().Id
	req.BindId = id
	return req, nil
}
