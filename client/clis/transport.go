package clis

import (
	"github.com/RussellLuo/timingwheel"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"time"
)

var timerMap = make(map[int32]*timingwheel.Timer)

// Transport manages client connections and request tracking.
type Transport struct {
	// client is the network connection.
	client *Client

	host string

	port int

	config *configs.ClientConfig
}

// NewTransport initializes a new Transport instance with the provided client configuration.
// It sets up the server host, port, and client configuration.
// Parameters:
//   - config: The client configuration to use.
//
// Returns:
//   - *Transport: A pointer to the newly created Transport instance.
func NewTransport(config *configs.ClientConfig) *Transport {
	// start reconnection.
	return &Transport{
		host:   config.ServerHost,
		port:   config.ServerPort,
		config: config,
	}
}

func (t *Transport) Connection(opts ...ClientOption) {
	t.client = NewClient(t.host, t.port)
	err := t.client.Connection("tcp", opts...)
	t.client.AddHandler(&CheckHandler{
		transport: t,
	})
	//The error add to reconnection list.
	if err != nil {
		log.Warn("Connection to server error:%s", err)
		addChecking(t)
	} else {
		t.openTunnel()
	}
}

func (t *Transport) openTunnel() {
	if t.client.isSmux() && t.config.Tunnels != nil {
		for _, cfg := range t.config.Tunnels {
			if err := t.client.OpenTunnel(cfg); err != nil {
				log.Warn("Connection to server error:%s %v", cfg.Type, err)
			}
		}
	}
}

func (t *Transport) SyncWrite(message exchange.InBound, timeout time.Duration) (*exchange.Protocol, error) {
	return SyncWrite(message, timeout, func(bytes []byte) error {
		return t.client.cct.Write(bytes)
	})
}

type ClientScheduler struct {
}

func (t *ClientScheduler) Next(t2 time.Time) time.Time {
	return t2.Add(3 * time.Second)
}

type CheckHandler struct {
	BaseClientHandler
	transport *Transport
}

func (b *CheckHandler) Close(cct *ClientControl) {
	addChecking(b.transport)
}

func (b *CheckHandler) Read(r *exchange.Protocol, cct *ClientControl) error {
	//Heart info.
	if r.Cmd == exchange.Heart {
		log.Debug("Receiver PONG info: %S", cct.cli.getAddress())
		return nil
	} else {
		Tracker.Complete(r.ReqId, r)
		return nil
	}
}

func (b *CheckHandler) Timeout(cct *ClientControl) {
	var h = exchange.Heartbeat{
		Value: "PING",
	}
	request, _ := exchange.NewRequest(h)
	_ = cct.Write(request.Bytes())
}

func checking(tp *Transport) {
	cli := tp.client
	if !cli.IsConnection() {
		log.Warn("Connection %s Not Active, start reconnection.", cli.getAddress())
		err := cli.doConnection()
		if err != nil {
			log.Warn("Reconnection %s Fail, next time still running.", cli.getAddress())
		} else {
			log.Info("👍<--Reconnection %s success OK.✅-->", cli.getAddress())
			tp.openTunnel()
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

func addChecking(tp *Transport) {
	if _, ok := timerMap[tp.client.id]; ok {
		return
	}
	t := utils.NewWheel.ScheduleFunc(&ClientScheduler{}, func() {
		checking(tp)
	})
	timerMap[tp.client.id] = t
}
