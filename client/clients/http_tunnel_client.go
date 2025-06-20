package clients

import (
	"bufio"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	"github.com/xtaci/smux"
	"io"
	"net"
	"net/http"
)

func init() {
	srv.RegisterTunnelClient("http", func(config *configs.ClientTunnelConfig) srv.TunnelClient {
		client := HttpTunnelClient{
			BaseTunnelClient: srv.BaseTunnelClient{
				Cfg: config,
			},
		}
		return &client
	})
}

type HttpTunnelClient struct {
	srv.BaseTunnelClient
	rw io.ReadWriteCloser
}

func (h *HttpTunnelClient) GetName() string {
	return "HttpTunnelClient"
}

func (h *HttpTunnelClient) Open(session *smux.Session) error {
	rw, err := session.OpenStream()
	if err != nil {
		log.Error("Open session fail %v", err)
		return err
	} else {
		log.Info("Open session success, %v:%v:%v ", h.GetName(), rw.ID(), rw.RemoteAddr())
		h.Register(rw)
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
		request.Write(dial)
		response, _ := http.ReadResponse(bufio.NewReader(dial), request)
		response.Write(rw)
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
