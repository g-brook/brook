package remote

import (
	"common/configs"
	"common/log"
	"github.com/RussellLuo/timingwheel"
	"time"
)

var tryList = make(map[int32]*Client)

var wheel *timingwheel.TimingWheel

type Transport struct {

	//
	//  clients
	//  @Description: all client.
	//
	clients []*Client

	ct int

	host string

	port int32

	config *configs.ClientConfig
}

// NewTransport
//
//	@Description: Init Transport.
//	@param ct
//	@return Transport
func NewTransport(ct int, config *configs.ClientConfig) *Transport {
	//reconnection.
	wheel = timingwheel.NewTimingWheel(time.Millisecond, 100)
	//start reconnection.
	return &Transport{
		ct:      ct,
		config:  config,
		clients: make([]*Client, 0),
	}
}

func (t *Transport) Connection(host string, port int32) {
	t.host = host
	t.port = port
	if t.ct <= 0 {
		t.ct = 1
	}
	for i := 0; i < t.ct; i++ {
		t.clients = append(t.clients, NewClient(host, port))
	}
	for _, cli := range t.clients {
		err := cli.Connection("tcp")
		//The error add to reconnection list.
		if err != nil {
			tryList[cli.id] = cli
			log.Warn("Connection to server error:%s", err)
		} else {
			//config := smux.DefaultConfig()
			//session, _ := smux.Client(cli.conn, config)
			//stream, err := session.OpenStream()
		}
	}
}
