package clis

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/google/uuid"
	"github.com/xtaci/smux"
)

type TunnelClientControl struct {
	Readers chan *exchange.Protocol

	Writers chan *exchange.Protocol

	Die chan struct{}
}

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
	Open(session *smux.Session) error

	//
	// Close
	//  @Description: Close
	//  @param session
	//
	Close()
}

// BaseTunnelClient is base impl.
type BaseTunnelClient struct {
	Stream *smux.Stream

	Cfg *configs.ClientTunnelConfig
}

func (b *BaseTunnelClient) GetName() string {
	return "BaseTunnelClient"
}

func (b *BaseTunnelClient) Open(_ *smux.Session) error {
	return nil
}

func (b *BaseTunnelClient) Close() {

}

func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReq {
	return exchange.RegisterReq{
		BindId:     uuid.New().String(),
		TunnelPort: b.Cfg.RemotePort,
	}
}

func (b *BaseTunnelClient) Register(stream *smux.Stream) {
	b.Stream = stream
	req := b.GetRegisterReq()
	request, _ := exchange.NewRequest(req)
	_, _ = b.Stream.Write(request.Bytes())
}

// at all tunnel tunnel by map.
var tunnels = make(map[string]FactoryFun)

// FactoryFun New tunnel client.
type FactoryFun func(config *configs.ClientTunnelConfig) TunnelClient

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
func GetTunnelClient(name string, config *configs.ClientTunnelConfig) TunnelClient {
	fun := tunnels[name]
	if fun != nil {
		return fun(config)
	}
	return nil
}

func GetTunnelClients() map[string]FactoryFun {
	return tunnels
}
