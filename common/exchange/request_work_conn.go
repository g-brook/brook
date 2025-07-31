package exchange

import "github.com/brook/common/utils"

type ReqWorkConn struct {
	Port         int              `json:"port"`
	ProxyId      string           `json:"proxy_id"`
	LocalAddress string           `json:"local_address"`
	TunnelType   utils.TunnelType `json:"tunnel_type"`
	UnId         string           `json:"un_id"`
}

func (r *ReqWorkConn) Cmd() Cmd {
	return WorkerConnReq
}
