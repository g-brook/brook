package tcp

import (
	"errors"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
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
// This function opens a tunnel server based on the request parameters.
func OpenTunnelServer(request exchange.OpenTunnelReq) (int, error) {
	//Check if the tunnel port is already in use.
	if _, ok := tcpTunnelServers.Load(request.TunnelPort); ok {
		//Lock the port.
		portPool.lockPort(request.TunnelPort)
		return request.TunnelPort, nil
	}
	//Check if the tunnel type is TCP.
	if request.TunnelType != utils.Tcp {
		return 0, fmt.Errorf("not tcp tunnel type, you need open tunnel type is %v", request.TunnelType)
	}
	//Get a dynamic port.
	if request.TunnelPort < portRange[0] {
		//Allocate a port from the pool.
		allocate, err := portPool.Allocate()
		if err != nil {
			return 0, err
		}
		request.TunnelPort = allocate
	} else {
		//Lock the port.
		isLock := portPool.lockPort(request.TunnelPort)
		//Check if the port is already in use.
		if !isLock {
			return 0, fmt.Errorf("the port: %d already use bind, Open tunnel server error, Pls change the port and try again", request.TunnelPort)
		}
	}
	//Create a new server configuration.
	config := &configs.ServerTunnelConfig{
		Port: request.TunnelPort,
		Type: utils.Tcp,
	}
	//Create a new base tunnel server.
	baseServer := tunnel.NewBaseTunnelServer(config)
	//Create a new TCP tunnel server.
	server := NewTcpTunnelServer(baseServer)
	//Start the server.
	err := server.Start()
	if err != nil {
		//Release the port if the server fails to start.
		portPool.Release(request.TunnelPort)
		return 0, err
	}
	//Store the server in the map.
	tcpTunnelServers.Store(request.TunnelPort, server)
	return request.TunnelPort, nil
}

// CloseTunnelServer This function closes a TCP tunnel server given a port number
func CloseTunnelServer(port int) {
	// Delete the TCP tunnel server from the map
	tcpTunnelServers.Delete(port)
	// Release the port back to the pool
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
