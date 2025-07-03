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

	closeConn := func(conn net.Conn) {
		if conn != nil {
			_ = conn.Close()
		}
	}
	call := func(request *http.Request, err error) (rsp *http.Response, dial net.Conn) {
		if err != nil {
			return
		}
		dial, err = net.Dial("tcp", h.GetCfg().LocalAddress)
		if err != nil {
			rsp = getErrorResponse()
			return
		}
		err = request.Write(dial)
		if err != nil {
			rsp = getErrorResponse()
			return
		}
		rsp, err = http.ReadResponse(bufio.NewReader(dial), request)
		if err != nil {
			rsp = getErrorResponse()
			return
		}
		return
	}
	loopRead := func() error {
		request, err := http.ReadRequest(bufio.NewReader(rw))
		response, dial := call(request, err)
		defer closeConn(dial)
		if response != nil && request != nil {
			requestId := request.Header.Get(RequestInfoKey)
			response.Header.Set(RequestInfoKey, requestId)
			_ = response.Write(rw)
			return nil
		} else {
			log.Warn("Read request fail", err)
			return err
		}
	}
	for {
		select {
		case <-h.Tcc.Context().Done():
			return
		default:

		}
		err := loopRead()
		if err == io.EOF {
			h.Close()
		}
	}

}

func getErrorResponse() *http.Response {
	return utils.GetResponse(http.StatusBadGateway)

}
