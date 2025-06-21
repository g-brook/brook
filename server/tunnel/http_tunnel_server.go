package tunnel

type HttpTunnelServer struct {
	*BaseTunnelServer
}

// NewHttpTunnelServer is a constructor function for HttpTunnelServer. It takes a pointer to BaseTunnelServer as input
// and returns a pointer to HttpTunnelServer. The constructor sets the DoStart field of BaseTunnelServer to the after
// method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations after the server
// processes the request. The constructor also returns a pointer to HttpTunnelServer.
func NewHttpTunnelServer(server *BaseTunnelServer) *HttpTunnelServer {
	tunnelServer := HttpTunnelServer{
		BaseTunnelServer: server,
	}
	server.DoStart = tunnelServer.after
	return &tunnelServer
}

// After is a method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations after
// the server processes the request.This method currently does not perform any operation, and returns nil directly.
// This may be a reserved hook point for future additions.Parameters:
// None Return value: error, indicating the result of the execution of the operation, and always returns nil.
func (tunnelServer *HttpTunnelServer) after() error {
	return nil
}
