package tcp

import (
	"errors"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/brook/server/tunnel"
	"math/rand"
	"sync"
	"time"
)

// TcpTunnelListener tcp tunnel listener

var tcpTunnelServers sync.Map

var portRange = [2]int{10000, 60000}

var portPool *PortPool

func init() {
	//tcpTunnelServers = make(map[int]*TcpTunnelServer, 50000)
	portPool = NewProtPool(portRange[0], portRange[1], time.Minute*5)
}

// OpenTunnelServer open tcp tunnel server
func OpenTunnelServer(ch transport.Channel, request exchange.RegisterReqAndRsp) (int, error) {
	if request.TunnelType != utils.Tcp {
		return 0, fmt.Errorf("not tcp tunnel type, you need open tunnel type is %v", request.TunnelType)
	}
	//Get a dynamic port.
	if request.TunnelPort < portRange[0] {
		allocate, err := portPool.Allocate()
		if err != nil {
			return 0, err
		}
		request.TunnelPort = allocate
	} else {
		isLock := portPool.lockPort(request.TunnelPort)
		if !isLock {
			return 0, fmt.Errorf("the port: %d already use bind, Open tunnel server error, Pls change the port and try again", request.TunnelPort)
		}
	}
	config := &configs.ServerTunnelConfig{
		Port: request.TunnelPort,
		Type: utils.Tcp,
	}
	baseServer := tunnel.NewBaseTunnelServer(config)
	server := NewTcpTunnelServer(baseServer)
	err := server.Start()
	if err != nil {
		portPool.Release(request.TunnelPort)
		return 0, err
	}
	server.RegisterConn(ch, request)
	//tcpTunnelServers[request.TunnelPort] = server
	tcpTunnelServers.Store(request.TunnelPort, server)
	return request.TunnelPort, nil
}

func CloseTunnelServer(port int) {
	tcpTunnelServers.Delete(port)
	portPool.Release(port)
}

type PortPool struct {
	mu      sync.Mutex
	ports   map[int]time.Time
	ttl     time.Duration
	minPort int
	maxPort int
}

func NewProtPool(minPort, maxPort int, ttl time.Duration) *PortPool {
	return &PortPool{
		minPort: minPort,
		maxPort: maxPort,
		ttl:     ttl,
		ports:   make(map[int]time.Time),
	}
}

func (p *PortPool) lockPort(port int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.ports[port]; exists {
		return false
	}
	p.ports[port] = time.Now()
	return true
}

// Allocate a port by unuse.
func (p *PortPool) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < (p.maxPort - p.minPort); i++ {
		port := rand.Intn(p.maxPort-p.minPort) + p.minPort
		if _, exists := p.ports[port]; !exists {
			p.ports[port] = time.Now()
			return port, nil
		}
	}
	return 0, errors.New("no available ports")
}

func (p *PortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.ports, port)
}
