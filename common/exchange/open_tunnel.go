package exchange

import "github.com/brook/common/utils"

type OpenTunnelReq struct {
	ProxyId      string           `json:"proxy_id"`
	TunnelType   utils.TunnelType `json:"type"`
	TunnelPort   int              `json:"port"`
	UnId         string           `json:"unId"`
	LocalAddress string           `json:"localAddress"`
}

func (o OpenTunnelReq) Cmd() Cmd {
	return OpenTunnel
}

type OpenTunnelResp struct {
	TunnelPort int    `json:"port"`
	UnId       string `json:"unId"`
}

func (o OpenTunnelResp) Cmd() Cmd {
	return OpenTunnel
}
