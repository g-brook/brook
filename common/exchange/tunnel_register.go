package exchange

import "github.com/brook/common/utils"

type TRegister interface {
	Cmd() Cmd

	GetTunnelType() utils.TunnelType

	GetProxyId() string

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
	ProxyId string `json:"proxy_id"`
}

func (r RegisterReqAndRsp) GetTunnelPort() int {
	return r.TunnelPort
}

func (r RegisterReqAndRsp) GetBindId() string {
	return r.BindId
}

func (r RegisterReqAndRsp) GetProxyId() string {
	return r.ProxyId
}

func (r RegisterReqAndRsp) GetTunnelType() utils.TunnelType {
	return r.TunnelType
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
