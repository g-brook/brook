package main

import (
	"bufio"
	"bytes"
	run "client/run"
	"common/configs"
	"common/log"
	"common/remote"
	"encoding/json"
	"errors"
	"fmt"
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
	//req := remote.RegisterReq{
	//	TunnelPort: 8818,
	//	BindId:     s,
	//}
	//by, _ := json.Marshal(req)
	//request := remote.NewRequest(remote.Register, by)
	//dial.Write(remote.Encoder(request))
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
	ch := make(chan remote.Protocol)
	dial, err := net.Dial("tcp", "127.0.0.1:8909")
	go reader(dial, ch)
	if err != nil {
		return "", errors.New("connection error")
	}
	registerRequest := remote.CommunicationInfo{}
	bytes, _ := json.Marshal(registerRequest)
	request := remote.NewRequest(remote.Communication, bytes)
	byts := remote.Encoder(request)
	_, _ = dial.Write(byts)
	m := <-ch
	if m.RspCode == remote.Rsp_success {
		fmt.Println("建立通道成功.")
		_ = json.Unmarshal(m.Data, &registerRequest)
		return registerRequest.BindId, nil
	}
	return "", errors.New("bind error")

}

func reader(conn net.Conn, ch chan remote.Protocol) {
	// 从服务器读取一行消息
	for true {
		reader := bufio.NewReader(conn)
		decoder, err := remote.Decoder(reader)
		if err != nil {
			fmt.Println("读取失败:", err)
			return
		}
		if decoder.Cmd == remote.Communication {
			ch <- decoder
		} else {
			fmt.Println("收到消息,", decoder.RspCode)
		}
	}

}
