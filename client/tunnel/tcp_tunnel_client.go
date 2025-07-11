package tunnel

import (
	"github.com/brook/client/clis"
	"github.com/brook/common/aio"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
	"io"
	"net"
	"time"
)

type ProxyConnection struct {
	io.ReadWriteCloser
}

func (proxy *ProxyConnection) Close() error {
	return nil
}

type TcpTunnelClient struct {
	*clis.BaseTunnelClient
	reconnect *clis.ReconnectManager
}

func init() {
	clis.RegisterTunnelClient(utils.Tcp, func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		tunnelClient := clis.NewBaseTunnelClient(config)
		client := TcpTunnelClient{
			BaseTunnelClient: tunnelClient,
		}
		tunnelClient.DoOpen = client.initOpen
		client.reconnect = clis.NewReconnectionManager(3 * time.Second)
		return &client
	})
}

func (t *TcpTunnelClient) GetName() string {
	return "tcp"
}

func (t *TcpTunnelClient) initOpen(_ *smux.Stream) error {
	t.BaseTunnelClient.AddReadHandler(exchange.WorkerConnReq, t.bindHandler)
	rsp, err := t.Register()
	if err != nil {
		log.Error("Register fail %v", err)
	} else {
		log.Info("Register success:PORT-%v", rsp.TunnelPort)
	}
	return nil
}

func (t *TcpTunnelClient) bindHandler(_ *exchange.Protocol, rw io.ReadWriteCloser) {
	dstCh := make(chan net.Conn)
	for {
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
		workConn := &ProxyConnection{
			rw,
		}
		errors := aio.Pipe(workConn, conn)
		if len(errors) > 0 {
			log.Error("Pipe error %v", errors)
		}
		_, err = t.checkSrcConnection(rw)
		if err != nil {
			rw.Close()
			break
		}
	}
}

func (t *TcpTunnelClient) checkSrcConnection(rw io.ReadWriteCloser) (int, error) {
	bytes := make([]byte, 0)
	return rw.Write(bytes)
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
