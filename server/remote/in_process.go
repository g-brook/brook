package remote

import (
	"fmt"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	defin "github.com/brook/server/define"
)

func init() {
	Register(exchange.Heart, pingProcess)
	Register(exchange.Register, registerProcess)
	Register(exchange.QueryTunnel, queryTunnelConfigProcess)
	Register(exchange.OpenTunnel, openTunnelProcess)
}

type InProcess[T exchange.InBound] func(request T, conn transport.Channel) (any, error)

// pingProcess
//
//	@Description:
//	@param request
//	@param conn
//	@return any
//	@return error
func pingProcess(request exchange.Heartbeat, conn transport.Channel) (any, error) {
	log.Debug("Receiver Ping message : %s:%v", request.Value, conn.RemoteAddr())
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
func registerProcess(request exchange.RegisterReq, conn transport.Channel) (any, error) {
	port := request.TunnelPort
	tunnel := defin.GetTunnel(port)
	if tunnel == nil {
		log.Error("Not found tunnel: %d", port)
		return nil, fmt.Errorf("not found tunnel:%d", port)
	}
	log.Debug("Registering tunnel:%v", tunnel)
	//tunnel.RegisterConn(conn, request)
	//Register conn to tunnel success.
	//conn.GetContext().AddAttr(isTunnelConnKey, true)
	//conn.GetContext().AddAttr(tunnelPort, port)
	return nil, nil
}

// queryTunnelConfigProcess
//
//	@Description: Query tunnel port config.
//	@param req
//	@param conn
func queryTunnelConfigProcess(req exchange.QueryTunnelReq, conn transport.Channel) (any, error) {
	tport := defin.Get[int](defin.TunnelPortKey)
	return exchange.QueryTunnelResp{
		TunnelPort: tport,
	}, nil
}

func openTunnelProcess(req exchange.OpenTunnelReq, conn transport.Channel) (any, error) {
	return exchange.OpenTunnelResp{
		SessionId: req.SessionId,
	}, nil
}
