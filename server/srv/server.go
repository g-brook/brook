package srv

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/brook/common/aio"
	"github.com/brook/common/hash"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/panjf2000/gnet/v2"
	"github.com/xtaci/smux"
)

var connLock sync.Mutex

type TraverseBy func()

type InitConnHandler func(conn *GChannel)

type ServerHandler interface {
	//
	// Close
	//  @Description: Shutdown conn notify.
	//  @param conn
	//
	Close(ch trp.Channel, traverse TraverseBy)

	//
	// Open
	//  @Description: Open conn notify.
	//  @param conn
	//
	Open(ch trp.Channel, traverse TraverseBy)

	//
	// Reader
	//  @Description: Reader conn data notify.
	//  @param conn
	//
	Reader(ch trp.Channel, traverse TraverseBy)

	//
	// Writer
	//  @Description: Writer data to conn.
	//  @param conn
	//  @param traverse
	//	// Writer
	Writer(ch trp.Channel, traverse TraverseBy)

	//
	// Boot
	//  @Description:
	//  @param s
	//  @param traverse
	//
	Boot(s *Server, traverse TraverseBy)
}

type BaseServerHandler struct {
}

func (b *BaseServerHandler) Writer(_ trp.Channel, traverse TraverseBy) {
	traverse()
}

func (b *BaseServerHandler) Close(_ trp.Channel, traverse TraverseBy) {
	traverse()
}

func (b *BaseServerHandler) Open(_ trp.Channel, traverse TraverseBy) {
	traverse()
}

func (b *BaseServerHandler) Reader(_ trp.Channel, traverse TraverseBy) {
	traverse()
}
func (b *BaseServerHandler) Boot(s *Server, traverse TraverseBy) {
	traverse()
}

func NewChannel(conn gnet.Conn, t *Server) *GChannel {
	ctx := conn.Context()
	if ctx == nil && t.isDatagram() {
		ctx = NewConnContext(t.isDatagram(), conn.RemoteAddr().String())
	}
	connContext := ctx.(*ConnContext)
	value, ok := t.connections.Load(connContext.Id)
	if ok {
		return value
	}
	connLock.Lock()
	defer connLock.Unlock()
	bgCtx, cancelFunc := context.WithCancel(context.Background())
	v2 := &GChannel{
		Conn:        conn,
		Id:          connContext.Id,
		Context:     connContext,
		Server:      t,
		bgCtx:       bgCtx,
		cancel:      cancelFunc,
		protocol:    t.opts.network,
		closeEvents: make([]trp.CloseEvent, 0),
	}
	if !t.isDatagram() {
		t.connections.Store(connContext.Id, v2)
	}
	if t.InitConnHandler != nil {
		t.InitConnHandler(v2)
	}
	return v2
}

// Server /*
type Server struct {
	*gnet.BuiltinEventEngine

	engine gnet.Engine

	// port.
	port int

	opts *sOptions

	connections *hash.SyncMap[string, *GChannel]

	handlers []ServerHandler

	InitConnHandler InitConnHandler

	startSmux func(conn *trp.SmuxAdapterConn, ctx context.Context, option *SmuxServerOption) error
}

func NewServer(port int) *Server {
	return &Server{
		port:        port,
		connections: hash.NewSyncMap[string, *GChannel](),
		handlers:    make([]ServerHandler, 0),
	}
}

func (sever *Server) AddHandler(handler ...ServerHandler) {
	sever.handlers = append(sever.handlers, handler...)
}

func (sever *Server) AddInitConnHandler(init InitConnHandler) {
	sever.InitConnHandler = init
}

func (sever *Server) Connections() map[string]*GChannel {
	tb := make(map[string]*GChannel)
	f := func(key string, value *GChannel) bool {
		tb[key] = value
		return true
	}
	sever.connections.Range(f)
	return tb
}

func (sever *Server) GetConnection(id string) (*GChannel, bool) {
	if sever.isDatagram() {
		log.Warn("server protocol is udp, can not get connection by id: %s", id)
		return nil, false
	}
	v2, ok := sever.connections.Load(id)
	return v2, ok
}
func (sever *Server) OnBoot(engine gnet.Engine) (action gnet.Action) {
	sever.engine = engine
	log.Info("Server started %d", sever.port)
	sever.next(func(s ServerHandler, conn trp.Channel) bool {
		b := true
		s.Boot(sever, func() {
			b = false
		})
		return b
	}, nil)
	return gnet.None
}

func (sever *Server) OnClose(c gnet.Conn, _ error) gnet.Action {
	log.Debug("Close an Connection: %s", c.RemoteAddr().String())
	conn2 := NewChannel(c, sever)
	conn2.GetContext().IsClosed = true
	defer sever.removeIfConnection(conn2)
	sever.next(func(s ServerHandler, newCh trp.Channel) bool {
		b := true
		s.Close(newCh, func() {
			b = false
		})
		return b
	}, conn2)
	return gnet.None
}

func (sever *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Debug("Open an Connection: %s", c.RemoteAddr().String())
	c.SetContext(NewConnContext(false, ""))
	conn := NewChannel(c, sever)
	defer sever.removeIfConnection(conn)
	if sever.startSmux != nil {
		conn.Context.isSmux = true
		conn.PipeConn = trp.NewSmuxAdapterConn(c)
		_ = sever.startSmux(
			conn.PipeConn,
			conn.bgCtx,
			nil,
		)
	} else {
		sever.next(func(s ServerHandler, newCh trp.Channel) bool {
			b := true
			s.Open(newCh, func() {
				b = false
			})
			return b
		}, conn)
	}
	return
}

func (sever *Server) OnTraffic(c gnet.Conn) gnet.Action {
	conn2 := NewChannel(c, sever)
	//Call lastActive.
	conn2.GetContext().LastActive()
	defer sever.removeIfConnection(conn2)
	if sever.startSmux != nil {
		if conn2.PipeConn != nil {
			buf, _ := conn2.Next(-1)
			_, err := conn2.PipeConn.Copy(buf)
			if err != nil {
				log.Error("pipeConn.Copy error: %s", err)
			}
		}
	} else {
		sever.next(func(s ServerHandler, newCh trp.Channel) bool {
			b := true
			s.Reader(newCh, func() {
				b = false
			})
			return b
		}, conn2)
	}
	return gnet.None
}

func (sever *Server) next(fun func(s ServerHandler, conn trp.Channel) bool, conn *GChannel) {
	for i := 0; i < len(sever.handlers); i++ {
		var newCh trp.Channel
		channelFunc := sever.opts.newChannelFunc
		if channelFunc != nil && conn != nil {
			newCh = channelFunc(conn)
		} else {
			newCh = conn
		}
		b := fun(sever.handlers[i], newCh)
		if b {
			break
		}
	}
}

// isDatagram checks if the server is using UDP protocol and initializes the event engine accordingly
func (sever *Server) isDatagram() bool {
	// Check if the server is configured to use UDP network
	return sever.opts.network == utils.NetworkUdp
}

func (sever *Server) GetPort() int {
	return sever.port
}

func (sever *Server) removeIfConnection(v2 *GChannel) {
	if !v2.isConnection() {
		//This use v2.id removing map element.
		// v2.id eq context.id, so yet use v2.id.
		//Because v2.context possible is nil.
		_, ok := sever.connections.Load(v2.Id)
		if ok {
			sever.connections.Delete(v2.Id)
		}
		_ = v2.Close()
	}
}

// Start is function start tcp server.
func (sever *Server) Start(opt ...ServerOption) error {
	//load sOptions config.
	sever.opts = serverOptions(opt...)
	if sever.opts.withSmux != nil && sever.opts.withSmux.enable {
		sever.startSmux = func(conn *trp.SmuxAdapterConn, ctx context.Context, option *SmuxServerOption) error {
			config := smux.DefaultConfig()
			session, err := smux.Server(conn, config)
			if err != nil {
				log.Error("Start server error. %v", err)
			}
			go func() {
				for {
					stream, _ := session.AcceptStream()
					if session.IsClosed() {
						log.Error("session is close.")
						return
					}
					log.Info("Start server success stream. %s", stream.RemoteAddr())
					channel := trp.NewSChannel(stream, ctx, false)
					sever.next(func(s ServerHandler, _ trp.Channel) bool {
						b := true
						s.Open(channel, func() {
							b = false
						})
						return b
					}, nil)
					go sever.smuxReadLoop(channel)
				}

			}()
			return nil
		}
	}
	network := sever.opts.network
	if network == "" {
		network = utils.NetworkTcp
		sever.opts.network = network
	}
	err := gnet.Run(sever, fmt.Sprintf("%s://:%d", network, sever.port),
		gnet.WithMulticore(true),
		gnet.WithLogger(&log.GnetLogger{}),
		gnet.WithReadBufferCap(65535),
		gnet.WithWriteBufferCap(65535),
		gnet.WithReusePort(true),
		gnet.WithReuseAddr(true),
	)
	if err != nil {
		log.Error("Error %v", err)
		return err
	}
	return nil
}

func (sever *Server) smuxReadLoop(ch *trp.SChannel) {
	for {
		if ch.IsTunnel {
			return
		}
		err := aio.WithBuffer(func(buf []byte) error {
			n, err := ch.Stream.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Error("stream is closed. %v", err)
					return err
				}
				log.Error("smux read error. %v", err)
				return nil
			}
			_, _ = ch.Copy(buf[:n])
			if !ch.IsTunnel {
				// If already this channel is bind tunnel,
				sever.next(func(s ServerHandler, _ trp.Channel) bool {
					b := true
					s.Reader(ch, func() {
						b = false
					})
					return b
				}, nil)
			}
			return nil
		}, aio.GetBytePool4k())
		if err != nil {
			return
		}
	}
}

func (sever *Server) Shutdown(ctx context.Context) {
	log.Info("Server shutdown: %d.", sever.GetPort())
	_ = sever.engine.Stop(ctx)
}
