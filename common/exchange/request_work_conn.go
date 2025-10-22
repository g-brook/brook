package exchange

type WorkConnReqByServer struct {
	ProxyId    string `json:"proxy_id"`
	RemotePort int    `json:"remote_port"`
}

func (r *WorkConnReqByServer) Cmd() Cmd {
	return WorkerConnReq
}
