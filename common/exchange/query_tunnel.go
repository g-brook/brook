package exchange

type QueryTunnelReq struct {
}

// Cmd
//
//	@Description: getCmd
//	@receiver r
//	@return Cmd
func (r QueryTunnelReq) Cmd() Cmd {
	return QueryTunnel
}

// QueryTunnelResp
// @Description: Resp.
type QueryTunnelResp struct {
	TunnelPort int32 `json:"tunnel_port"`
}

func (r QueryTunnelReq) QueryTunnelResp() Cmd {
	return QueryTunnel
}
