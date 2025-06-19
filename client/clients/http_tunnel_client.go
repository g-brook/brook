package clients

import (
	"bufio"
	"fmt"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	"github.com/xtaci/smux"
	"io"
	"net"
	"net/http"
	"time"
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
	rw, err := session.OpenStream()
	if err != nil {
		log.Error("Open session fail", err)
		return
	} else {
		fmt.Println("Open session success", rw.ID())
		_, err := rw.Write([]byte("PING"))
		fmt.Println(err)
	}
	h.rw = rw
	go h.bindHandler(rw)
	for {
		time.Sleep(3 * time.Second)
		h.rw.Write([]byte("ping"))
		fmt.Println("发送一个数据.....")
	}
}

func (h *HttpTunnelClient) bindHandler(rw io.ReadWriteCloser) {
	loopRead := func() error {
		request, err := http.ReadRequest(bufio.NewReader(rw))
		if err != nil {
			log.Error("Read request fail", err.Error())
			rw.Write([]byte("ping"))
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
