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
		ProxyId:    m.cfg.ProxyId,
		TunnelType: utils.Tcp,
		TunnelPort: m.cfg.RemotePort,
	}
	// Send the request to the server and wait for a response.
	_, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		// Log an error if the request fails.
		log.Error("Open tcp tunnel server error %v", err)
		return err
	}
	// Check if a TCP client already exists for the given proxy ID.
	if _, ok := m.tcpClients.Load(m.cfg.ProxyId); !ok {
		// If not, create a new TCP client.
		client := NewTcpTunnelClient(m.cfg, m)
		// Open the TCP client.
		err = client.Open(session)
		if err != nil {
			// Log an error if the TCP client fails to open.
			log.Error("Open tcp tunnel server error %v", err)
			return err
		}
		// Store the TCP client in the map.
		m.tcpClients.Store(m.cfg.ProxyId, client)
	}
	// Log that the TCP tunnel server was opened successfully.
	log.Info("Open tcp tunnel server success")
	return nil
}

func (m *MultipleTunnelClient) Close() {

}
