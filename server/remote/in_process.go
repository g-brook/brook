package remote

import (
	"fmt"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	defin "github.com/brook/server/define"
	"github.com/brook/server/srv"
)

func init() {
	Register(exchange.Heart, pingProcess)
	Register(exchange.Register, registerProcess)
	Register(exchange.QueryTunnel, queryTunnelConfigProcess)
	Register(exchange.OpenTunnel, openTunnelProcess)
}

type InProcess[T exchange.InBound] func(request T, ch transport.Channel) (any, error)

// pingProcess
//
//	@Description:
//	@param request
//	@param ch
//	@return any
//	@return error
func pingProcess(request exchange.Heartbeat, ch transport.Channel) (any, error) {
	log.Debug("Receiver Ping message : %s:%v", request.Value, ch.RemoteAddr())
	heartbeat := exchange.Heartbeat{Value: "PONG"}
	return heartbeat, nil
}

// registerProcess
//
//	@Description:
//	@param request
//	@param ch
//	@return any
//	@return error
func registerProcess(request exchange.RegisterReqAndRsp, ch transport.Channel) (any, error) {
	port := request.TunnelPort
	tunnel := srv.GetTunnel(port)
	if tunnel == nil {
		log.Error("Not found tunnel: %d", port)
		return nil, fmt.Errorf("not found tunnel:%d", port)
	}
	log.Debug("Registering tunnel:%v", tunnel)
	return request, nil
}

// queryTunnelConfigProcess
//
//	@Description: Query tunnel port config.
//	@param req
//	@param ch
func queryTunnelConfigProcess(req exchange.QueryTunnelReq, _ transport.Channel) (any, error) {
	port := defin.Get[int](defin.TunnelPortKey)
	return exchange.QueryTunnelResp{
		TunnelPort: port,
	}, nil
}

func openTunnelProcess(req exchange.OpenTunnelReq, _ transport.Channel) (any, error) {
	return exchange.OpenTunnelResp{
		SessionId: req.SessionId,
	}, nil
}
