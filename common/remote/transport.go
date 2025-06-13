package remote

import (
	"fmt"
	"github.com/RussellLuo/timingwheel"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
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

func init() {
	newWheel = timingwheel.NewTimingWheel(time.Millisecond, 100)
	newWheel.Start()
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
		t.clients = append(t.clients, NewClient(t.host, t.port, &ClientControl{
			state: make(chan ClientState),
		}))
	}
	for _, cli := range t.clients {
		go func() {
			err := cli.Connection("tcp", opts...)
			//The error add to reconnection list.
			if err != nil {
				tryList[cli.id] = cli
				log.Warn("Connection to server error:%s", err)
			} else {
				wheel.ScheduleFunc(&ClientScheduler{
					client: cli,
				}, func() {
					fmt.Println("true")
				})
			}
		}()
		select {
		case state := <-cli.cct.state:
			fmt.Println(state, "xxx.......x")
		}
	}
}

type ClientScheduler struct {
	client *Client
}

func (t *ClientScheduler) Next(t2 time.Time) time.Time {
	return t2.Add(3000 * time.Millisecond)
}
