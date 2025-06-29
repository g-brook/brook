package tunnel

import (
	"context"
	"github.com/brook/client/clis"
	"github.com/brook/common"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/xtaci/smux"
	"io"
	"time"
)

func init() {
	clis.RegisterTunnelClient(common.EchoTest, func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		return &EchoTunnelClient{
			tcc: clis.NewTunnelClientControl(context.Background()),
		}
	})
}

type EchoTunnelClient struct {
	tcc     *clis.TunnelClientControl
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
		tcc := e.tcc
		tcc.Bucket = exchange.NewMessageBucket(rw, e.tcc.Context())
		tcc.Bucket.AddHandler(exchange.Heart, e.pingRev)
		tcc.Bucket.Run()
		go e.sendPing()
	}
	return nil
}

func (e *EchoTunnelClient) Close() {
	if e.session != nil {
		_ = e.session.Close()
	}
	e.tcc.Cancel()
}

func (e *EchoTunnelClient) Done() <-chan struct{} {
	return e.tcc.Context().Done()
}

func (e *EchoTunnelClient) pingRev(_ *exchange.Protocol, _ io.ReadWriteCloser) {
	log.Debug("%v Receive Pong.....", e.session.RemoteAddr())
}

func (e *EchoTunnelClient) sendPing() {
	timePing := time.NewTimer(time.Second * 5)
	defer timePing.Stop()
	for {
		select {
		case <-e.tcc.Context().Done():
			return
		case <-timePing.C:
			heartbeat := exchange.Heartbeat{
				Value: "PING",
			}
			request, _ := exchange.NewRequest(heartbeat)
			err := e.tcc.Bucket.Push(request)
			if err != nil {
				return
			}
			timePing.Reset(time.Second * 5)
		}
	}
}
