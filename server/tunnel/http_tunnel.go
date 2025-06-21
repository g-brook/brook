package tunnel

//
//import (
//	"errors"
//	"fmt"
//	"github.com/brook/common/configs"
//	"github.com/brook/common/exchange"
//	"github.com/brook/common/log"
//	"github.com/brook/common/transport"
//	server "github.com/brook/server/remote"
//	srv2 "github.com/brook/server/srv"
//	"io"
//	"net"
//	"net/http"
//	"net/http/httputil"
//	"sync"
//)
//
//type TcpListener struct {
//	conns chan net.Conn
//}
//
//func NewTcpListener() *TcpListener {
//	return &TcpListener{conns: make(chan net.Conn, 128)}
//}
//
//func (t *TcpListener) Accept() (net.Conn, error) {
//	if conn, ok := <-t.conns; ok {
//		return conn, nil
//	}
//	return nil, errors.New("listener close")
//}
//
//func (t *TcpListener) Close() error {
//	fmt.Println("Close")
//	return nil
//}
//
//func (t *TcpListener) Addr() net.Addr {
//	return (*net.TCPAddr)(nil)
//}
//
//func (t *TcpListener) PutConn(conn net.Conn) {
//	t.conns <- conn
//}
//
//type HttpTunnel struct {
//	srv2.BaseServerHandler
//	config      *configs.TunnelConfig
//	server      *server.InServer
//	tl          *TcpListener
//	tc          sync.WaitGroup
//	refChannels map[string]transport.Channel
//
//	fromChannels map[string]transport.Channel
//}
//
//func (h *HttpTunnel) Open(conn transport.Channel, traverse srv2.TraverseBy) {
//	//h.tl.PutConn(conn)
//}
//
//func (h *HttpTunnel) Boot(conn *srv2.Server, traverse srv2.TraverseBy) {
//	h.tc.Done()
//}
//
//func (h *HttpTunnel) Reader(conn transport.Channel, traverse srv2.TraverseBy) {
//	length := len(h.refChannels)
//	if length > 0 {
//		var keys = make([]string, 0, length)
//		for key := range h.refChannels {
//			keys = append(keys, key)
//		}
//		firstKey := keys[0]
//		target := h.refChannels[firstKey]
//		h.fromChannels[firstKey] = conn
//		_, err := io.Copy(target.GetWriter(), conn.GetReader())
//		if err != nil {
//			log.Warn("Error....")
//		}
//	}
//	traverse()
//}
//
//func NewHttpTunnel(config *configs.TunnelConfig, server *server.InServer) *HttpTunnel {
//	return &HttpTunnel{config: config, server: server, tc: sync.WaitGroup{}, refChannels: make(map[string]transport.Channel), fromChannels: make(map[string]transport.Channel)}
//}
//
//func (h *HttpTunnel) Start() {
//	h.tc.Add(1)
//	go func() {
//		newServer := srv2.NewServer(h.Port())
//		newServer.AddHandler(h)
//		srv2.AddTunnel(h)
//		err := newServer.Start()
//		if err != nil {
//			log.Error("Started Http newServer error: %s", h.Port())
//		}
//	}()
//	go func() {
//		h.tc.Wait()
//		if h.server == nil {
//			log.Warn("Server is nil")
//			return
//		}
//		h.tl = NewTcpListener()
//		log.Info("Started Http tunnel success %d", h.Port())
//		rp := &httputil.ReverseProxy{
//			Rewrite: func(request *httputil.ProxyRequest) {
//				out := request.Out
//				out.URL.Scheme = "http"
//			},
//		}
//		server := http.Server{Handler: rp, ReadHeaderTimeout: 0}
//		err := server.Serve(h.tl)
//		if err != nil {
//			log.Info("HttpTunnel server stop")
//		}
//	}()
//}
//
//func (h *HttpTunnel) Port() int {
//	return h.config.Port
//}
//
//func (h *HttpTunnel) RegisterConn(v2 *srv2.GChannel, request exchange.RegisterReqAndRsp) {
//	//t.refChannels = append(t.refChannels, v2)
//	h.refChannels[v2.GetContext().Id] = v2
//	log.Info("Bind tcp tunnel conn t(tunnel/server): %d c(client): %d", h.Port(), v2.RemoteAddr())
//}
//
//func (h *HttpTunnel) Receiver(conn *srv2.GChannel) {
//	id := conn.GetContext().Id
//	toConn, ok := h.fromChannels[id]
//	if ok {
//		_, err := io.Copy(toConn.GetWriter(), conn.GetReader())
//		if err != nil {
//			log.Error("Copy to transport fail %v", err)
//		}
//	} else {
//		log.Warn("Not found tunnel conn,%s", id)
//	}
//}
