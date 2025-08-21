package remote

import (
	"fmt"
	"time"

	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	defin "github.com/brook/server/define"
	"github.com/brook/server/tunnel"
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
	heartbeat := exchange.Heartbeat{Value: "PONG",
		StartTime:  request.StartTime,
		ServerTime: time.Now().UnixMilli(),
	}
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
	switch sch := ch.(type) {
	case *transport.SChannel:
		sch.IsTunnel = true
		sch.AddAttr(defin.TunnelProxyId, request.ProxyId)
	default:
		log.Error("Not support channel type: %T", ch)
		return nil, fmt.Errorf("not support channel type:%T", ch)
	}
	request.BindId = ch.GetId()
	port := request.TunnelPort
	t := tunnel.GetTunnel(port)
	if t == nil {
		log.Error("Not found tunnel: %d", port)
		return nil, fmt.Errorf("not found tunnel:%d", port)
	}
	log.Debug("Registering tunnel:%v", t)
	t.RegisterConn(ch, request)
	return request, nil
}

// queryTunnelConfigProcess
//
//	@Description: Query tunnel port config.
//	@param req
//	@param ch
func queryTunnelConfigProcess(req exchange.QueryTunnelReq, ch transport.Channel) (any, error) {
	port := defin.Get[int](defin.TunnelPortKey)
	return exchange.QueryTunnelResp{
		TunnelPort: port,
		UnId:       ch.GetId(),
	}, nil
}

func openTunnelProcess(req exchange.OpenTunnelReq, ch transport.Channel) (any, error) {
	openPort, err := OpenTunnelServer(req, ch)
	if err != nil {
		return nil, err
	}
	return exchange.OpenTunnelResp{
		UnId:       req.UnId,
		TunnelPort: openPort,
	}, nil
}
