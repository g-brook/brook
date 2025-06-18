package clients

import (
	"bufio"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	"github.com/xtaci/smux"
	"io"
	"net"
	"net/http"
)

func init() {
	srv.RegisterTunnelClient("http", func() srv.TunnelClient {
		return &HttpTunnelClient{}
	})
}

type HttpTunnelClient struct {
	rw io.ReadWriteCloser
}

func (h *HttpTunnelClient) GetName() string {
	return "HttpTunnelClient"
}

func (h *HttpTunnelClient) Open(session *smux.Session) {
	if session == nil {
		log.Error("Open session is nil")
		return
	}
	rw, err := session.Open()
	if err != nil {
		log.Error("Open session fail")
		return
	}
	h.rw = rw
	h.bindHandler(rw)
}

func (h *HttpTunnelClient) bindHandler(rw io.ReadWriteCloser) {
	loopRead := func() {
		request, err := http.ReadRequest(bufio.NewReader(rw))
		if err != nil {
			log.Error("Read request fail")
			return
		}
		dial, err := net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			return
		}
		defer dial.Close()
		request.Write(dial)
		response, _ := http.ReadResponse(bufio.NewReader(dial), request)
		response.Write(rw)
	}
	for {
		loopRead()
	}
}

type HttpWriter struct {
	writer io.Writer

	reader io.Reader
}
