package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brook/client/run"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	"io"
	"net"
	"net/http"
)

func init() {
	log.InitFunc(configs.LoggerConfig{LogPath: "./", LoggLevel: "info"})
}
func main() {
	run.Start()
	//s, err := communication()
	//if err != nil {
	//	return
	//}
	////this is tunnel connection.
	//dial, _ := net.Dial("tcp", "127.0.0.1:8909")
	//go copy(dial)
	////
	//req := srv.RegisterReq{
	//	TunnelPort: 8818,
	//	BindId:     s,
	//}
	//by, _ := json.Marshal(req)
	//request := srv.NewRequest(srv.Register, by)
	//dial.Write(srv.Encoder(request))
	//time.Sleep(36000 * time.Second)
}

type HttpWriter struct {
	conn net.Conn
}

func (h HttpWriter) Write(p []byte) (n int, err error) {
	//conn2, err := net.Dial("tcp", "127.0.0.1:8080")
	//fmt.Println("收到数据了。。。。。。。。")
	//go func() {
	//	fmt.Println("发送数据了。")
	//	defer conn2.Close()
	//	//io.Copy(conn2, h.conn)
	//}()
	//io.Copy(h.conn, conn2)
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(p)))
	if err != nil {
		return
	}
	conn2, err := net.Dial("tcp", "127.0.0.1:8080")
	defer conn2.Close()
	request.Write(conn2)
	//conn2.Write(p)
	resp, err := http.ReadResponse(bufio.NewReader(conn2), request)
	resp.Write(h.conn)
	return len(p), nil
}

func copy(dial net.Conn) {
	var httpWriter = HttpWriter{
		conn: dial,
	}
	_, err := io.Copy(&httpWriter, dial)
	if err != nil {
		return
	}
}

func communication() (string, error) {
	ch := make(chan srv.Protocol)
	dial, err := net.Dial("tcp", "127.0.0.1:8909")
	if err != nil {
		return "", errors.New("connection error")
	}
	registerRequest := srv.CommunicationInfo{}
	bytes, _ := json.Marshal(registerRequest)
	request := srv.NewRequest(srv.Communication, bytes)
	byts := srv.Encoder(request)
	_, _ = dial.Write(byts)
	m := <-ch
	if m.RspCode == srv.RspSuccess {
		fmt.Println("建立通道成功.")
		_ = json.Unmarshal(m.Data, &registerRequest)
		return registerRequest.BindId, nil
	}
	return "", errors.New("bind error")

}
