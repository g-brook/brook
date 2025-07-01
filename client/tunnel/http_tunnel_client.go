package tunnel

import (
	"bufio"
	"io"
	"net"
	"net/http"

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
)

var (
	RequestInfoKey = "httpRequestId"
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
	loopRead := func() error {
		request, err := http.ReadRequest(bufio.NewReader(rw))
		if err != nil {
			writeError(rw)
			return err
		}
		dial, err := net.Dial("tcp", h.GetCfg().LocalAddress)

		//close.
		defer func(dial net.Conn) {
			_ = dial.Close()
		}(dial)

		if err != nil {
			writeError(rw)
			return err
		}
		requestId := request.Header.Get(RequestInfoKey)
		err = request.Write(dial)
		if err != nil {
			writeError(rw)
			return err
		}
		response, err := http.ReadResponse(bufio.NewReader(dial), request)
		if err != nil {
			writeError(rw)
			return err
		}
		response.Header.Set(RequestInfoKey, requestId)
		err = response.Write(rw)
		if err != nil {
			return err
		}
		return nil
	}
	for {
		_ = loopRead()
	}

}

func writeError(rw io.ReadWriter) {
	response := utils.GetResponse(http.StatusBadGateway)
	_ = response.Write(rw)
}
