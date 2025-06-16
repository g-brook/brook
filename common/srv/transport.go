package srv

import (
	"encoding/json"
	"github.com/RussellLuo/timingwheel"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"time"
)

var timerMap = make(map[int32]*timingwheel.Timer)

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
	//start reconnection.
	return &Transport{
		clients: make([]*Client, 0),
		ct:      ct,
		host:    config.ServerHost,
		port:    config.ServerPort,
		config:  config,
	}
}

func (t *Transport) Connection(opts ...ClientOption) {
	for i := 0; i < t.ct; i++ {
		t.clients = append(t.clients, NewClient(t.host, t.port))
	}
	for _, cli := range t.clients {
		go func() {
			err := cli.Connection("tcp", opts...)
			cli.AddHandler(&CheckHandler{})
			//The error add to reconnection list.
			if err != nil {
				log.Warn("Connection to server error:%s", err)
				addChecking(cli)
			}
		}()
	}
}

type ClientScheduler struct {
}

func (t *ClientScheduler) Next(t2 time.Time) time.Time {
	return t2.Add(3000 * time.Millisecond)
}

type CheckHandler struct {
	BaseClientHandler
}

func (e *CheckHandler) Close(cct *ClientControl) {
	addChecking(cct.cli)
}

func (b *CheckHandler) Read(_ *Protocol, cct *ClientControl) (int, error) {
	log.Debug("Receiver PONG info: %S", cct.cli.getAddress())
	return 0, nil
}

func (e *CheckHandler) Timeout(cct *ClientControl) {
	var h = Heartbeat{
		Value: "PING",
	}
	bytes, _ := json.Marshal(h)
	request := NewRequest(Heart, bytes)
	b := Encoder(request)
	cct.Write(b)
}

func checking(cli *Client) {
	if !cli.IsConnection() {
		log.Warn("Connection %s Not Active, start reconnection.", cli.getAddress())
		err := cli.doConnection()
		if err != nil {
			log.Warn("Reconnection %s Fail, next time still running.", cli.getAddress())
		} else {
			log.Info("👍<--Reconnection %s success OK.✅-->", cli.getAddress())
		}
	}
	defer func() {
		if cli.IsConnection() {
			timer, ok := timerMap[cli.id]
			if ok {
				timer.Stop()
				delete(timerMap, cli.id)
			}
		}
	}()
}

func addChecking(cli *Client) {
	if _, ok := timerMap[cli.id]; ok {
		return
	}
	t := newWheel.ScheduleFunc(&ClientScheduler{}, func() {
		checking(cli)
	})
	timerMap[cli.id] = t
}
