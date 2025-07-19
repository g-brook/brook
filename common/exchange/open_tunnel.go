package exchange

import "github.com/brook/common/utils"

type OpenTunnelReq struct {
	SessionId  string           `json:"sessionId"`
	ProxyId    string           `json:"proxy_id"`
	TunnelType utils.TunnelType `json:"type"`
	TunnelPort int              `json:"port"`
}

func (o OpenTunnelReq) Cmd() Cmd {
	return OpenTunnel
}

type OpenTunnelResp struct {
	SessionId  string `json:"sessionId"`
	TunnelPort int    `json:"port"`
}

func (o OpenTunnelResp) Cmd() Cmd {
	return OpenTunnel
}
