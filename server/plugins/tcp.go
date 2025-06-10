package plugins

import "common/remote"

type TcpServerPlugin struct {
}

func (s TcpServerPlugin) Name() string {
	return "tcp"
}

func (s TcpServerPlugin) Run(client remote.Server) {

}
