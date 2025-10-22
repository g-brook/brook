package exchange

type LoginReq struct {
	Token string `json:"token"`
}

// Cmd
//
//	@Description: getCmd
//	@receiver r
//	@return Cmd
func (r LoginReq) Cmd() Cmd {
	return LoginTunnel
}

// LoginResp
// @Description: Resp.
type LoginResp struct {
	TunnelPort int `json:"tunnel_port"`

	UnId string `json:"un_id"`

	AutoTunnel bool `json:"auto_login"`
}

func (r LoginReq) QueryTunnelResp() Cmd {
	return LoginTunnel
}
