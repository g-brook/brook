package tunnel

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/brook/client/clis"
	"github.com/brook/common/aio"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
)

type TcpTunnelClient struct {
	*clis.BaseTunnelClient
	reconnect   *clis.ReconnectManager
	multipleTcp *MultipleTunnelClient
}

func NewTcpTunnelClient(config *configs.ClientTunnelConfig, mtpc *MultipleTunnelClient) *TcpTunnelClient {
	tunnelClient := clis.NewBaseTunnelClient(config, false)
	client := TcpTunnelClient{
		BaseTunnelClient: tunnelClient,
		multipleTcp:      mtpc,
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	client.reconnect = clis.NewReconnectionManager(3 * time.Second)
	return &client
}

func (t *TcpTunnelClient) GetName() string {
	return "TcpTunnelClient"
}

func (t *TcpTunnelClient) initOpen(ch *transport.SChannel) error {
	localConnection, err := t.localConnection()
	if err != nil {
		if localConnection != nil {
			_ = localConnection.Close()
		}
		_ = ch.Close()
		return err
	}
	err = t.AsyncRegister(t.GetRegisterReq(), func(p *exchange.Protocol, rw io.ReadWriteCloser) error {
		log.Info("Connection local address success then Client to server register success:%v", t.GetCfg().LocalAddress)
		if p.IsSuccess() {
			errors := aio.Pipe(ch, localConnection)
			if len(errors) > 0 {
				log.Error("Pipe error %v", errors)
			}
			return nil
		} else {
			log.Error("Connection local address success then Client to server register fail:%v", t.GetCfg().LocalAddress)
			return fmt.Errorf("register fail")
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
func (t *TcpTunnelClient) localConnection() (net.Conn, error) {
	connFunction := func() (net.Conn, error) {
		dial, err := net.Dial(string(utils.NetworkTcp), t.GetCfg().LocalAddress)
		if err != nil {
			return nil, err
		}
		log.Info("Connection localAddress, %v success", t.GetCfg().LocalAddress)
		return dial, err
	}
	return connFunction()
}
