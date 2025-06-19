package srv

import (
	"fmt"
	"github.com/brook/common/log"
	"github.com/panjf2000/gnet/v2"
	"github.com/xtaci/smux"
	"sync"
)

var connLock sync.Mutex

type TraverseBy func()

type InitConnHandler func(conn *GChannel)

type ServerHandler interface {
	//
	// Close
	//  @Description: Close conn notify.
	//  @param conn
	//
	Close(ch Channel, traverse TraverseBy)

	//
	// Open
	//  @Description: Open conn notify.
	//  @param conn
	//
	Open(ch Channel, traverse TraverseBy)

	//
	// Reader
	//  @Description: Reader conn data notify.
	//  @param conn
	//
	Reader(ch Channel, traverse TraverseBy)

	//
	// Writer
	//  @Description: Writer data to conn.
	//  @param conn
	//  @param traverse
	//	// Writer
	Writer(ch Channel, traverse TraverseBy)

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

func (b BaseServerHandler) Writer(ch Channel, traverse TraverseBy) {
	traverse()
}

func (b BaseServerHandler) Close(ch Channel, traverse TraverseBy) {
	traverse()
}

func (b BaseServerHandler) Open(ch Channel, traverse TraverseBy) {
	traverse()
}

func (b BaseServerHandler) Reader(ch Channel, traverse TraverseBy) {
	traverse()
}
func (b BaseServerHandler) Boot(s *Server, traverse TraverseBy) {
	traverse()
}

func newConn2(conn gnet.Conn, t *Server) *GChannel {
	ctx := conn.Context()
	context := ctx.(*ConnContext)
	value, ok := t.connections[context.Id]
	if ok {
		return value
	}
	connLock.Lock()
	defer connLock.Unlock()
	v2 := &GChannel{
		conn:    conn,
		id:      context.Id,
		context: context,
		server:  t,
	}
	t.connections[context.Id] = v2
	if t.InitConnHandler != nil {
		t.InitConnHandler(v2)
	}
	return v2
}

// Server /*
type Server struct {
	*gnet.BuiltinEventEngine

	// port.
	port int32

	opts *sOptions

	connections map[string]*GChannel

	handlers []ServerHandler

	InitConnHandler InitConnHandler

	startSmux func(conn *SmuxAdapterConn, option *SmuxServerOption) error
}

func NewServer(port int32) *Server {
	return &Server{
		port:        port,
		handlers:    make([]ServerHandler, 0),
		connections: make(map[string]*GChannel),
	}
}

func (sever *Server) AddHandler(handler ...ServerHandler) {
	sever.handlers = append(sever.handlers, handler...)
}

func (sever *Server) AddInitConnHandler(init InitConnHandler) {
	sever.InitConnHandler = init
}

func (sever *Server) Connections() map[string]*GChannel {
	return sever.connections
}

func (sever *Server) GetConnection(id string) (*GChannel, bool) {
	v2, ok := sever.connections[id]
	return v2, ok
}

func (sever *Server) OnBoot(_ gnet.Engine) (action gnet.Action) {
	log.Info("Server started %d", sever.port)
	sever.next(func(s ServerHandler) bool {
		b := true
		s.Boot(sever, func() {
			b = false
		})
		return b
	})
	return gnet.None
}

func (sever *Server) OnClose(c gnet.Conn, _ error) gnet.Action {
	log.Info("Close an Connection: %s", c.RemoteAddr().String())
	conn2 := newConn2(c, sever)
	defer sever.removeIfConnection(conn2)
	sever.next(func(s ServerHandler) bool {
		b := true
		s.Close(conn2, func() {
			b = false
		})
		return b
	})
	return gnet.None
}

func (sever *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Info("Open an Connection: %s", c.RemoteAddr().String())
	c.SetContext(NewConnContext())
	conn2 := newConn2(c, sever)
	defer sever.removeIfConnection(conn2)
	if sever.startSmux != nil {
		conn2.context.isSmux = true
		conn2.pipeConn = NewSmuxAdapterConn(c)
		_ = sever.startSmux(conn2.pipeConn, nil)
	} else {
		sever.next(func(s ServerHandler) bool {
			b := true
			s.Open(conn2, func() {
				b = false
			})
			return b
		})
	}
	return
}

func (sever *Server) OnTraffic(c gnet.Conn) gnet.Action {
	conn2 := newConn2(c, sever)
	//Call lastActive.
	conn2.GetContext().LastActive()
	defer sever.removeIfConnection(conn2)
	if sever.startSmux != nil {
		if conn2.pipeConn != nil {
			buf, _ := conn2.Next(-1)
			_, err := conn2.pipeConn.Copy(buf)
			if err != nil {
				log.Error("pipeConn.Copy error: %s", err)
			}
		}
	} else {
		sever.next(func(s ServerHandler) bool {
			b := true
			//Call Reader method.
			s.Reader(conn2, func() {
				b = false
			})
			return b
		})
	}
	return gnet.None
}

func (sever *Server) next(fun func(s ServerHandler) bool) {
	for i := 0; i < len(sever.handlers); i++ {
		b := fun(sever.handlers[i])
		if b {
			break
		}
	}
}

func (sever *Server) GetPort() int32 {
	return sever.port
}

func (sever *Server) removeIfConnection(v2 *GChannel) {
	if !v2.isConnection() {
		//This use v2.id removing map element.
		// v2.id eq context.id, so yet use v2.id.
		//Because v2.context possible is nil.
		delete(sever.connections, v2.id)
		_ = v2.Close()
	}
}

// Start is function start tcp server.
func (sever *Server) Start(opt ...ServerOption) error {
	//load sOptions config.
	sever.opts = serverOptions(opt...)
	if sever.opts.withSmux != nil && sever.opts.withSmux.enable {
		sever.startSmux = func(conn *SmuxAdapterConn, option *SmuxServerOption) error {
			go func() {
				config := smux.DefaultConfig()
				session, err := smux.Server(conn, config)
				if err != nil {
					log.Error("Start server error.", err)
					return
				}
				stream, err := session.AcceptStream()
				if err != nil {
					log.Error("Start server error.", err)
					return
				}
				fmt.Println("Open session success", stream.ID())
				sever.next(func(s ServerHandler) bool {
					b := true
					s.Reader(NewSChannel(stream), func() {
						b = false
					})
					return b
				})
				return
			}()
			return nil
		}
	}
	err := gnet.Run(sever, fmt.Sprintf("tcp://:%d", sever.port),
		gnet.WithMulticore(true),
		gnet.WithLogger(&log.GnetLogger{}),
		gnet.WithReadBufferCap(65535),
		gnet.WithWriteBufferCap(65535),
		gnet.WithReusePort(true),
	)
	if err != nil {
		log.Error("Error", err)
		return err
	}
	return nil
}

func (sever *Server) Close() {
	log.Info("Server close.")
}
