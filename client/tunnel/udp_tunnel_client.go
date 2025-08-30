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
	"github.com/brook/common/utils"
)

type UdpTunnelClient struct {
	*clis.BaseTunnelClient
	reconnect    *clis.ReconnectManager
	multipleTcp  *MultipleTunnelClient
	localAddress *net.UDPAddr
	bufSize      int
}

func NewUdpTunnelClient(cfg *configs.ClientTunnelConfig, mtpc *MultipleTunnelClient) *UdpTunnelClient {
	if cfg.UdpSize == 0 {
		cfg.UdpSize = 1500
	}
	tunnelClient := clis.NewBaseTunnelClient(cfg, false)
	client := UdpTunnelClient{
		BaseTunnelClient: tunnelClient,
		multipleTcp:      mtpc,
		bufSize:          cfg.UdpSize,
	}
	var err error
	client.localAddress, err = net.ResolveUDPAddr("udp", cfg.LocalAddress)
	if err != nil {
		log.Error("NewUdpTunnelClient error %v", err)
		return nil
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	client.reconnect = clis.NewReconnectionManager(3 * time.Second)
	return &client
}

func (t *UdpTunnelClient) GetName() string {
	return "udp"
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
	stop := make(chan struct{})
	revLoop := func(rw io.ReadWriteCloser, bucket *exchange.TunnelBucket) {
		bucket.DefaultRead(func(p *exchange.TunnelProtocol) {
			_, err2 := localConnection.Write(p.Data)
			if err2 != nil {
				log.Error("Write to local address error %v", err2)
				close(stop)
			}
		})
	}
	readLoop := func(rw *net.UDPConn, bucket *exchange.TunnelBucket) {
		pool := aio.GetByteBufPool(t.bufSize)
		for {
			err := aio.WithBuffer(func(buf []byte) error {
				_, _, err = rw.ReadFromUDP(buf)
				if err != nil {
					return err
				}
				_ = bucket.Push(buf, nil)
				return err
			}, pool)
			if err != nil && err == io.EOF {
				close(stop)
				return
			}
			select {
			case <-stop:
				return
			default:
			}
		}
	}
	err = t.AsyncRegister(func(p *exchange.Protocol, rw io.ReadWriteCloser) {
		log.Info("Connection local address success then Client to server register success:%v", t.GetCfg().LocalAddress)
		bucket := exchange.NewTunnelBucket(rw, t.Tcc.Context())
		go revLoop(rw, bucket)
		go readLoop(localConnection, bucket)
		<-stop
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
func (t *UdpTunnelClient) localConnection() (*net.UDPConn, error) {
	connFunction := func() (*net.UDPConn, error) {
		dial, err := net.DialUDP(string(utils.NetworkUdp), nil, t.localAddress)
		if err != nil {
			return nil, err
		}
		log.Info("Connection localAddress, %v success", t.GetCfg().LocalAddress)
		return dial, err
	}
	return connFunction()
}
