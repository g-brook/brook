package srv

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/transport"
)

// Save all tunnels channel. port: server.
var tunnels map[int]TunnelServer

func init() {
	tunnels = make(map[int]TunnelServer)
}

// AddTunnel
//
//	@Description: Add tunnel server.
//	@param tunnel
func AddTunnel(tunnel TunnelServer) {
	tunnels[tunnel.Port()] = tunnel
}

// GetTunnel
//
//	@Description: Get TunnelServer server.
//	@param port
//	@return TunnelServer
func GetTunnel(port int) TunnelServer {
	tunnel, ok := tunnels[port]
	if ok {
		return tunnel
	}
	return nil
}

// TunnelServer
// @Description: Define TunnelServer interface.
type TunnelServer interface {

	// Start is start tunnel server.
	Start(cfg *configs.ServerTunnelConfig) error

	//
	// Port
	//  @Description:  Get tunnel port.
	//  @return int32
	//
	Port() int

	//
	// RegisterConn
	//  @Description: Register connection to TunnelServer.
	//  @param v2 connection.
	//  @param request request.
	//
	RegisterConn(v2 *transport.SChannel, request exchange.RegisterReqAndRsp)

	//
	// Receiver
	//  @Description: copy data to tunnel.
	//  @param v2
	//
	Receiver(v2 *transport.SChannel)
}
