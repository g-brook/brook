package exchange

type OpenTunnelReq struct {
	ProxyId string `json:"proxy_id"`
	UnId    string `json:"unId"`
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
