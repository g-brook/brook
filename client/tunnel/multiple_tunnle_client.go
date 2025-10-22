package tunnel

import (
	"sync"
	"time"

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
)

// This function is used to register a new tunnel client for the TCP protocol
func init() {
	ft := func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		client := MultipleTunnelClient{
			clientTunnelConfig: config,
			tcpClients:         sync.Map{},
		}
		return &client
	}
	// Register the new tunnel client with the clis package
	clis.RegisterTunnelClient(utils.Tcp, ft)
	clis.RegisterTunnelClient(utils.Udp, ft)
}

type MultipleTunnelClient struct {
	clientTunnelConfig *configs.ClientTunnelConfig
	tcpClients         sync.Map
}

func (m *MultipleTunnelClient) GetName() string {
	return "multiple-tunnel-client"
}

// Open This function opens a TCP tunnel server for a given session.
func (m *MultipleTunnelClient) Open(session *smux.Session) error {
	// Create a new OpenTunnelReq struct with the proxy ID, tunnel type, and tunnel port.
	req := &exchange.OpenTunnelReq{
		ProxyId: m.clientTunnelConfig.ProxyId,
		UnId:    clis.ManagerTransport.UnId,
	}
	rsp, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		log.Error("Open %v tunnel server error %v", m.clientTunnelConfig.ProxyId, err)
		return err
	}
	if !rsp.IsSuccess() {
		log.Error("Open %v tunnel server error %v", m.clientTunnelConfig.ProxyId, rsp.RspMsg)
		return err
	}
	parse, _ := exchange.Parse[exchange.OpenTunnelResp](rsp.Data)
	clis.ManagerTransport.PutConfig(m.clientTunnelConfig)
	clis.ManagerTransport.AddMessage(exchange.WorkerConnReq, func(r *exchange.Protocol) error {
		go func() {
			reqWorder, _ := exchange.Parse[exchange.WorkConnReqByServer](r.Data)
			config := clis.ManagerTransport.GetConfig(reqWorder.ProxyId)
			if config == nil {
				return
			}
			config.RemotePort = reqWorder.RemotePort
			if config.TunnelType == utils.Tcp {
				client := NewTcpTunnelClient(config, m)
				_ = client.Open(session)
				_ = client.OpenStream()
			} else if config.TunnelType == utils.Udp {
				client := NewUdpTunnelClient(config, m)
				_ = client.Open(session)
				_ = client.OpenStream()
			}
		}()
		return nil
	})
	// Log that the TCP tunnel server was opened successfully.
	log.Info("Open %v tunnel client success:%v:%v", m.clientTunnelConfig.TunnelType, m.clientTunnelConfig.ProxyId, parse.TunnelPort)
	return nil
}

func (m *MultipleTunnelClient) Close() {

}
