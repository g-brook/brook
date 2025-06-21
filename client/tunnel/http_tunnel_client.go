package tunnel

import (
	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/xtaci/smux"
	"io"
)

func init() {
	clis.RegisterTunnelClient("http", func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		tunnelClient := clis.NewBaseTunnelClient(config)
		client := HttpTunnelClient{
			BaseTunnelClient: tunnelClient,
		}
		tunnelClient.DoOpen = client.initOpen
		return &client
	})
}

type HttpTunnelClient struct {
	*clis.BaseTunnelClient
	rw io.ReadWriteCloser
}

func (h *HttpTunnelClient) GetName() string {
	return "HttpTunnelClient"
}

func (h *HttpTunnelClient) initOpen(_ *smux.Stream) error {
	err := h.Register()
	if err != nil {
		log.Error("Register fail %v", err)
	} else {
		log.Info("Register success")
	}
	return nil
}

type HttpWriter struct {
	writer io.Writer

	reader io.Reader
}
