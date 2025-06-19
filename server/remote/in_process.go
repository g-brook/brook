package remote

import (
	"fmt"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	defin "github.com/brook/server/define"
)

func init() {
	Register(exchange.Heart, pingProcess)
	Register(exchange.Register, registerProcess)
	Register(exchange.Communication, communicationProcess)
	Register(exchange.QueryTunnel, queryTunnelConfigProcess)
	Register(exchange.OpenTunnel, openTunnelProcess)
}

type InProcess[T exchange.InBound] func(request T, conn *srv.GChannel) (any, error)

// pingProcess
//
//	@Description:
//	@param request
//	@param conn
//	@return any
//	@return error
func pingProcess(request exchange.Heartbeat, conn *srv.GChannel) (any, error) {
	log.Debug("Receiver Ping message : %s:%s", request.Value, conn.RemoteAddr().String())
	heartbeat := exchange.Heartbeat{Value: "PONG"}
	return heartbeat, nil
}

// registerProcess
//
//	@Description:
//	@param request
//	@param conn
//	@return any
//	@return error
func registerProcess(request exchange.RegisterReq, conn *srv.GChannel) (any, error) {
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

// communicationProcess
//
//	@Description:
//	@param req
//	@param conn
//	@return any
//	@return error
func communicationProcess(req exchange.CommunicationInfo, conn *srv.GChannel) (any, error) {
	id := conn.GetContext().Id
	req.BindId = id
	return req, nil
}

// queryTunnelConfigProcess
//
//	@Description: Query tunnel port config.
//	@param req
//	@param conn
func queryTunnelConfigProcess(req exchange.QueryTunnelReq, conn *srv.GChannel) (any, error) {
	tport := defin.Get[int32](defin.TunnelPortKey)
	return exchange.QueryTunnelResp{
		TunnelPort: tport,
	}, nil
}

func openTunnelProcess(req exchange.OpenTunnelReq, conn *srv.GChannel) (any, error) {
	return exchange.OpenTunnelResp{
		SessionId: req.SessionId,
	}, nil
}
