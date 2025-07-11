package exchange

import "github.com/brook/common/utils"

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

func (r RegisterReqAndRsp) Cmd() Cmd {
	return Register
}
