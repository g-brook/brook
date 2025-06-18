package defin

import (
	"github.com/brook/common/exchange"
	"github.com/brook/common/srv"
)

// Save all tunnels channel. port: server.
var tunnels map[int32]Tunnel

func init() {
	tunnels = make(map[int32]Tunnel)
}

// AddTunnel
//
//	@Description: Add tunnel server.
//	@param tunnel
func AddTunnel(tunnel Tunnel) {
	tunnels[tunnel.Port()] = tunnel
}

// GetTunnel
//
//	@Description: Get Tunnel server.
//	@param port
//	@return Tunnel
func GetTunnel(port int32) Tunnel {
	tunnel, ok := tunnels[port]
	if ok {
		return tunnel
	}
	return nil
}

// Tunnel
// @Description: Define Tunnel interface.
type Tunnel interface {
	//
	// Port
	//  @Description:  Get tunnel port.
	//  @return int32
	//
	Port() int32

	//
	// RegisterConn
	//  @Description: Register connection to Tunnel.
	//  @param v2 connection.
	//  @param request request.
	//
	RegisterConn(v2 *srv.ConnV2, request exchange.RegisterReq)

	//
	// Receiver
	//  @Description: copy data to tunnel.
	//  @param v2
	//
	Receiver(v2 *srv.ConnV2)
}
