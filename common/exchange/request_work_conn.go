package exchange

type WorkConnReq struct {
	ProxyId    string `json:"proxy_id"`
	RemotePort int    `json:"remote_port"`
}

func (r *WorkConnReq) Cmd() Cmd {
	return WorkerConnReq
}
