package tunnel

import (
	"github.com/brook/client/clis"
	"github.com/brook/common/aio"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"io"
	"net"
	"time"
)

type TcpTunnelClient struct {
	*clis.BaseTunnelClient
	reconnect  *clis.ReconnectManager
	closeTaget chan struct{}
	mtcp       *MultipleTunnelClient
}

func NewTcpTunnelClient(config *configs.ClientTunnelConfig, mtpc *MultipleTunnelClient) *TcpTunnelClient {
	tunnelClient := clis.NewBaseTunnelClient(config)
	client := TcpTunnelClient{
		BaseTunnelClient: tunnelClient,
		mtcp:             mtpc,
	}
	tunnelClient.DoOpen = client.initOpen
	client.reconnect = clis.NewReconnectionManager(3 * time.Second)
	return &client
}

func (t *TcpTunnelClient) GetName() string {
	return "tcp"
}

func (t *TcpTunnelClient) initOpen(_ *transport.SChannel) error {
	t.BaseTunnelClient.AddReadHandler(exchange.WorkerConnReq, t.bindHandler)
	rsp, err := t.Register()
	if err != nil {
		log.Error("Register fail %v", err)
		return err
	} else {
		log.Info("Register success:PORT-%v", rsp.TunnelPort)
	}
	return nil
}

func (t *TcpTunnelClient) bindHandler(_ *exchange.Protocol, rw io.ReadWriteCloser) {
	dstCh := make(chan net.Conn)
	conn, err := t.reconnection()
	if err != nil {
		t.reconnect.TryReconnect(func() bool {
			conn, err = t.reconnection()
			if err != nil {
				return false
			}
			dstCh <- conn
			return true
		})
		conn = <-dstCh
	}
	errors := aio.Pipe(rw, conn)
	if len(errors) > 0 {
		log.Error("Pipe error %v", errors)
	}
}

func (t *TcpTunnelClient) reconnection() (net.Conn, error) {
	connFunction := func() (net.Conn, error) {
		dial, err := net.Dial("tcp", t.GetCfg().LocalAddress)
		if err != nil {
			return nil, err
		}
		log.Info("Connection %v success", t.GetCfg().LocalAddress)
		return dial, err
	}
	return connFunction()
}
