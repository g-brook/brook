package exchange

import "github.com/brook/common/utils"

type TRegister interface {
	Cmd() Cmd

	GetTunnelType() utils.TunnelType

	GetProxyId() string

	GetHttpId() string

	GetTunnelPort() int

	GetBindId() string
}

// RegisterReqAndRsp
// @Description: Register Info.
type RegisterReqAndRsp struct {
	TunnelType utils.TunnelType `json:"tunnel_type"`

	// TunnelPort is port.
	TunnelPort int `json:"tunnel_port"`

	//request id.
	BindId string `json:"bind_id"`

	//proxy id. only http or http.
	HttpId string `json:"http_id"`

	//proxyId.
	ProxyId string `json:"proxyId"`
}

func (r RegisterReqAndRsp) GetTunnelPort() int {
	return r.TunnelPort
}

func (r RegisterReqAndRsp) GetBindId() string {
	return r.BindId
}

func (r RegisterReqAndRsp) GetHttpId() string {
	return r.HttpId
}

func (r RegisterReqAndRsp) GetTunnelType() utils.TunnelType {
	return r.TunnelType
}

func (r RegisterReqAndRsp) GetProxyId() string {
	return r.ProxyId
}

type UdpRegisterReqAndRsp struct {
	RegisterReqAndRsp
	RemoteAddress string `json:"remote_address"`
}

func (r UdpRegisterReqAndRsp) Cmd() Cmd {
	return UdpRegister
}

func (r RegisterReqAndRsp) Cmd() Cmd {
	return Register
}
