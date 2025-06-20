package srv

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
	"io"
	"time"
)

func init() {
	RegisterTunnelClient(utils.EchoTest, func(config *configs.ClientTunnelConfig) TunnelClient {
		return &EchoTunnelClient{
			tcc: &TunnelClientControl{
				readers: make(chan *exchange.Protocol, 1),
				writers: make(chan *exchange.Protocol, 1),
				die:     make(chan struct{}),
			},
		}
	})
}

type EchoTunnelClient struct {
	rw      io.ReadWriteCloser
	tcc     *TunnelClientControl
	session *smux.Session
}

func (e *EchoTunnelClient) GetName() string {
	return "EchoTunnel"
}

func (e *EchoTunnelClient) Open(session *smux.Session) error {
	e.session = session
	rw, err := session.Open()
	if err != nil {
		log.Error("Open session %s error: %v", e.GetName(), err)
		return err
	} else {
		log.Info("Open session %s success %v:%v", e.GetName(), session.NumStreams(), session.RemoteAddr())
		e.rw = rw
		go e.revLoop()
		go e.readLoop()
	}
	return nil
}

func (e *EchoTunnelClient) Close() {
	close(e.tcc.die)
	if e.session != nil {
		_ = e.session.Close()
	}
}

func (e *EchoTunnelClient) readLoop() {
	for {
		pr, err := exchange.Decoder(e.rw)
		if err != nil {
			if err == io.EOF {
				e.Close()
				return
			}
			log.Error("Decoder %s error: %v", e.GetName(), err)
		} else {
			if pr.Cmd == exchange.Heart {
				log.Debug("Receive Heart.....")
			}
		}
	}
}

func (e *EchoTunnelClient) revLoop() {
	timePing := time.NewTimer(time.Second * 5)
	defer timePing.Stop()
	for {
		select {
		case <-e.tcc.die:
			return
		case <-timePing.C:
			heartbeat := exchange.Heartbeat{
				Value: "PING",
			}
			request, _ := exchange.NewRequest(heartbeat)
			_, _ = e.rw.Write(request.Bytes())
			log.Debug("Ping heartbeat")
			timePing.Reset(time.Second * 5)
		}
	}
}
