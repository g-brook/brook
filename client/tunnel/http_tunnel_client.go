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
	"net/http"
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

// HttpTunnelClient is a tunnel client that handles HTTP connections.
type HttpTunnelClient struct {
	*clis.BaseTunnelClient
}

// GetName returns the name of the tunnel client.
func (h *HttpTunnelClient) GetName() string {
	return "HttpTunnelClient"
}

// initOpen initializes the HTTP tunnel client by registering it and logging the result.
// Parameters:
//   - stream: The smux stream to use.
//
// Returns:
//   - error: An error if the registration fails.
func (h *HttpTunnelClient) initOpen(_ *smux.Stream) error {
	h.BaseTunnelClient.AddRead(exchange.WorkerConnReq, h.bindHandler)
	err := h.Register()
	if err != nil {
		log.Error("Register fail %v", err)
	} else {
		log.Info("Register success")
	}
	return nil
}

func (h *HttpTunnelClient) bindHandler(_ *exchange.Protocol, rw io.ReadWriteCloser) {
	localConn, err := net.Dial("tcp", h.GetCfg().LocalAddress)
	if err != nil {
		log.Error("Connect %v", err)
		rw.Close()
		return
	}
	errors := aio.Pipe(rw, localConn)
	if len(errors) > 0 {
		log.Error("Pipe error")
	}
}

func writeError(rw io.ReadWriter) {
	response := utils.GetResponse(http.StatusBadGateway)
	_ = response.Write(rw)
}
