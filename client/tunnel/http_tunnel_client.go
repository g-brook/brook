package tunnel

import (
	"bufio"
	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
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
	rw io.ReadWriteCloser
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
	err := h.Register()
	if err != nil {
		log.Error("Register fail %v", err)
	} else {
		log.Info("Register success")
		go h.bindHandler(h.GetReaderWriter())
	}
	return nil
}

func (h *HttpTunnelClient) bindHandler(rw io.ReadWriteCloser) {
	loopRead := func() error {
		request, err := http.ReadRequest(bufio.NewReader(rw))
		if err != nil {
			log.Error("Read request fail", err.Error())
			return err
		}
		dial, err := net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			return nil
		}
		defer dial.Close()
		_ = request.Write(dial)
		response, _ := http.ReadResponse(bufio.NewReader(dial), request)
		_ = response.Write(rw)
		return nil
	}
	for {
		err := loopRead()
		if err != nil {
			break
		}
	}
}

type HttpWriter struct {
	writer io.Writer

	reader io.Reader
}
