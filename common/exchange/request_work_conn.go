package exchange

type ReqWorkConn struct {
}

func (r *ReqWorkConn) Cmd() Cmd {
	return WorkerConnReq
}
