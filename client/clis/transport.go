package clis

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"time"
)

// Transport
// @Description:Transport manages client and request tracking.
type Transport struct {

	// client　is net connection.
	client *Client

	host string

	port int

	config *configs.ClientConfig

	reconnect *ReconnectManager
}

// NewTransport
//
//	@Description: Init Transport.
//	@param ct
//	@return Transport
func NewTransport(config *configs.ClientConfig) *Transport {
	//start reconnection.
	return &Transport{
		host:      config.ServerHost,
		port:      config.ServerPort,
		config:    config,
		reconnect: NewReconnectionManager(time.Second * 5),
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
	return exchange.SyncWriteInBound(message, timeout, func(protocol *exchange.Protocol) error {
		return t.client.cct.Write(protocol.Bytes())
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
		log.Debug("Receiver PONG info: %v", cct.cli.getAddress())
		return nil
	} else {
		exchange.Tracker.Complete(r)
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

func addChecking(tp *Transport) {
	reconnect := func() bool {
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
		return cli.IsConnection()
	}
	tp.reconnect.TryReconnect(reconnect)
}
