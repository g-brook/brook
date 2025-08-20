package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
)

var (
	RequestInfoKey = "httpRequestId"
)

func init() {
	clis.RegisterTunnelClient(utils.Http, func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		tunnelClient := clis.NewBaseTunnelClient(config, true)
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
func (h *HttpTunnelClient) initOpen(sch *transport.SChannel) error {
	h.BaseTunnelClient.AddReadHandler(exchange.WorkerConnReq, h.bindHandler)
	rsp, err := h.Register()
	if err != nil {
		log.Error("Register fail %v", err)
		return err
	} else {
		log.Info("Register success:PORT-%v", rsp.TunnelPort)
	}
	return nil
}

func (h *HttpTunnelClient) bindHandler(req *exchange.Protocol, rw io.ReadWriteCloser) {
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
			bodyBytes, _ := io.ReadAll(response.Body)
			response.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
			headerBytes := BuildCustomHTTPHeader(response)
			merged := append(headerBytes, bodyBytes...)
			_, err = rw.Write(merged)
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

func BuildCustomHTTPHeader(r *http.Response) []byte {
	var buf bytes.Buffer

	st := fmt.Sprintf("HTTP/%d.%d %03d %s\r\n", r.ProtoMajor, r.ProtoMinor, r.StatusCode, r.Status)
	// 写入状态行（例如：HTTP/1.1 200 OK）
	buf.WriteString(st)

	// 写入所有 headers
	for key, value := range r.Header {
		if key != "Transfer-Encoding" {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value[0]))
		}
	}

	// 空行，结束 Header 区域
	buf.WriteString("\r\n")

	return buf.Bytes()
}

func getErrorResponse() *http.Response {
	return utils.GetResponse(http.StatusInternalServerError)

}
