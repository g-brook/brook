package exchange

type OpenTunnelReq struct {
	SessionId string `json:"sessionId"`
}

func (o OpenTunnelReq) Cmd() Cmd {
	return OpenTunnel
}

type OpenTunnelResp struct {
	SessionId string `json:"sessionId"`
}

func (o OpenTunnelResp) Cmd() Cmd {
	return OpenTunnel
}
