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
			cfg:        config,
			tcpClients: sync.Map{},
		}
		return &client
	}
	// Register the new tunnel client with the clis package
	clis.RegisterTunnelClient(utils.Tcp, ft)
	clis.RegisterTunnelClient(utils.Udp, ft)
}

type MultipleTunnelClient struct {
	cfg        *configs.ClientTunnelConfig
	tcpClients sync.Map
}

func (m *MultipleTunnelClient) GetName() string {
	return "multiple-tunnel-client"
}

// Open This function opens a TCP tunnel server for a given session.
func (m *MultipleTunnelClient) Open(session *smux.Session) error {
	// Create a new OpenTunnelReq struct with the proxy ID, tunnel type, and tunnel port.
	req := &exchange.OpenTunnelReq{
		ProxyId:      m.cfg.ProxyId,
		TunnelType:   m.cfg.TunnelType,
		TunnelPort:   m.cfg.RemotePort,
		UnId:         clis.ManagerTransport.UnId,
		LocalAddress: m.cfg.LocalAddress,
	}
	rsp, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		log.Error("Open %v tunnel server error %v:%v", req.TunnelType, req.TunnelPort, err)
		return err
	}
	if !rsp.IsSuccess() {
		log.Error("Open %v tunnel server error %v:%v", req.TunnelType, req.TunnelPort, rsp.RspMsg)
		return err
	}
	clis.ManagerTransport.AddMessage(exchange.WorkerConnReq, func(r *exchange.Protocol) error {
		go func() {
			reqWorder, _ := exchange.Parse[exchange.ReqWorkConn](r.Data)
			newCfg := &configs.ClientTunnelConfig{
				ProxyId:      reqWorder.ProxyId,
				RemotePort:   reqWorder.Port,
				LocalAddress: reqWorder.LocalAddress,
				TunnelType:   reqWorder.TunnelType,
			}

			if reqWorder.Network == utils.NetworkTcp {
				client := NewTcpTunnelClient(newCfg, m)
				_ = client.Open(session)
				_ = client.OpenStream()
			} else if reqWorder.Network == utils.NetworkUdp {
				client := NewUdpTunnelClient(newCfg, m)
				_ = client.Open(session)
				_ = client.OpenStream()
			}
		}()
		return nil
	})
	// Log that the TCP tunnel server was opened successfully.
	log.Info("Open %v tunnel client success:%v:%v", m.cfg.TunnelType, m.cfg.ProxyId, m.cfg.RemotePort)
	return nil
}

func (m *MultipleTunnelClient) Close() {

}
