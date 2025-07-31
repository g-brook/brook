package remote

import (
	"fmt"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
)

var Inserver *InServer

var OpenTunnelServerFun func(req exchange.OpenTunnelReq, ch transport.Channel) (int, error)

// InWriteMessage This function takes a user ID and an InBound request as parameters and sends the request to the user's connection
func InWriteMessage(unId string, r exchange.InBound) error {
	//Get the connection associated with the user ID
	connection, ok := Inserver.GetConnection(unId)
	//If the connection exists
	if ok {
		//Create a new request from the InBound request
		request, _ := exchange.NewRequest(r)
		//Write the request to the connection
		_, err := connection.Write(request.Bytes())
		return err
	} else {
		return fmt.Errorf("not found user %s connection", unId)
	}
}

func OpenTunnelServer(req exchange.OpenTunnelReq, ch transport.Channel) (int, error) {
	if OpenTunnelServerFun == nil {
		log.Error("not found open tunnel function")
		return 0, fmt.Errorf("not found open tunnel function")
	}
	return OpenTunnelServerFun(req, ch)
}
