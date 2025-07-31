package tunnel

import (
	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
	"sync"
	"time"
)

// This function is used to register a new tunnel client for the TCP protocol
func init() {
	// Register the new tunnel client with the clis package
	clis.RegisterTunnelClient(utils.Tcp, func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		// Create a new MultipleTunnelClient instance
		client := MultipleTunnelClient{
			// Set the configuration for the client
			cfg: config,
			// Initialize the tcpClients map
			tcpClients: sync.Map{},
		}
		// Return the new client instance
		return &client
	})
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
		TunnelType:   utils.Tcp,
		TunnelPort:   m.cfg.RemotePort,
		UnId:         clis.ManagerTransport.UnId,
		LocalAddress: m.cfg.LocalAddress,
	}
	// Send the request to the server and wait for a response.
	rsp, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		// Log an error if the request fails.
		log.Error("Open tcp tunnel server error %v:%v", req.TunnelPort, err)
		return err
	}
	if !rsp.IsSuccess() {
		log.Error("Open tcp tunnel server error %v:%v", req.TunnelPort, rsp.RspMsg)
		return err
	}
	clis.ManagerTransport.AddMessage(exchange.WorkerConnReq, func(r *exchange.Protocol) error {
		// If not, create a new TCP client.
		go func() {
			// Store the TCP client in the map.
			reqWorder, _ := exchange.Parse[exchange.ReqWorkConn](r.Data)
			newCfg := &configs.ClientTunnelConfig{
				ProxyId:      reqWorder.ProxyId,
				RemotePort:   reqWorder.Port,
				LocalAddress: reqWorder.LocalAddress,
				TunnelType:   reqWorder.TunnelType,
			}
			client := NewTcpTunnelClient(newCfg, m)
			_ = client.Open(session)
			_ = client.OpenStream()
		}()
		return nil
	})
	// Log that the TCP tunnel server was opened successfully.
	log.Info("Open tcp tunnel client success:%v:%v", m.cfg.ProxyId, m.cfg.RemotePort)
	return nil
}

func (m *MultipleTunnelClient) Close() {

}
