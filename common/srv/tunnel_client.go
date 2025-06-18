package srv

import "github.com/xtaci/smux"

type TunnelClient interface {

	//
	// GetName
	//  @Description: Get name.
	//  @return string
	//
	GetName() string

	//
	// Open
	//  @Description: Open tunnel.
	//  @param session
	//
	Open(session *smux.Session)
}

// at all tunnel clients by map.
var tunnels = make(map[string]FactoryFun)

// FactoryFun New tunnel client.
type FactoryFun func() TunnelClient

// RegisterTunnelClient
//
//	@Description: Register tunnel client.
//	@param name
//	@param factory
func RegisterTunnelClient(name string, factory FactoryFun) {
	tunnels[name] = factory
}

// GetTunnelClient
//
//	@Description: Get tunnel client.
//	@param name
//	@return TunnelClient
func GetTunnelClient(name string) TunnelClient {
	fun := tunnels[name]
	if fun != nil {
		return fun()
	}
	return nil
}
