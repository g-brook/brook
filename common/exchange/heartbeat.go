package exchange

// Heartbeat
// @Description: Ping InBound info. This is empty request,server use Cmd　discern.
type Heartbeat struct {
	Value      string `json:"value"`
	StartTime  int64  `json:"start_time"`
	ServerTime int64  `json:"server_time"`
}

// Cmd
//
//	@Description: getCmd.
//	@receiver p
//	@return Cmd
func (p Heartbeat) Cmd() Cmd {
	return Heart
}
