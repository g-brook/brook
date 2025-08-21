package tunnel

import (
	"io"
	"net"
	"time"

	"github.com/brook/client/clis"
	"github.com/brook/common/aio"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
)

type UdpTunnelClient struct {
	*clis.BaseTunnelClient
	reconnect   *clis.ReconnectManager
	multipleTcp *MultipleTunnelClient
}

func NewUdpTunnelClient(config *configs.ClientTunnelConfig, mtpc *MultipleTunnelClient) *UdpTunnelClient {
	tunnelClient := clis.NewBaseTunnelClient(config, false)
	client := UdpTunnelClient{
		BaseTunnelClient: tunnelClient,
		multipleTcp:      mtpc,
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	client.reconnect = clis.NewReconnectionManager(3 * time.Second)
	return &client
}

func (t *UdpTunnelClient) GetName() string {
	return "upd"
}

func (t *UdpTunnelClient) initOpen(ch *transport.SChannel) error {
	localConnection, err := t.localConnection()
	if err != nil {
		if localConnection != nil {
			_ = localConnection.Close()
		}
		_ = ch.Close()
		return err
	}
	err = t.AsyncRegister(func(p *exchange.Protocol, rw io.ReadWriteCloser) {
		log.Info("Connection local address success then Client to server register success:%v", t.GetCfg().LocalAddress)
		errors := aio.Pipe(ch, localConnection)
		if len(errors) > 0 {
			log.Error("Pipe error %v", errors)
		}
	})
	if err != nil {
		if localConnection != nil {
			_ = localConnection.Close()
		}
		_ = ch.Close()
		log.Error("Connection fail %v", err)
		return err
	}
	return nil
}
func (t *UdpTunnelClient) localConnection() (net.Conn, error) {
	connFunction := func() (net.Conn, error) {
		dial, err := net.Dial("udp", t.GetCfg().LocalAddress)
		if err != nil {
			return nil, err
		}
		log.Info("Connection localAddress, %v success", t.GetCfg().LocalAddress)
		return dial, err
	}
	return connFunction()
}
