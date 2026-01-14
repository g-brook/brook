/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"

	"github.com/quic-go/quic-go"
)

func generateTLSConfig() *tls.Config {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, _ := tls.X509KeyPair(certPEM, keyPEM)
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}

func main() {
	fmt.Println("hello world")

	addr, err := quic.ListenAddr(":7000", generateTLSConfig(), nil)
	if err != nil {
		panic(err)
	}
	n := addr.Addr()
	if n != nil {
		conn, err := addr.Accept(context.Background())
		if err != nil {
			panic(err)
		}
		go handleConnection(conn)
	}
	fmt.Println(n)
	<-context.Background().Done()
}

func handleConnection(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			fmt.Printf("AcceptStream error: %v\n", err)
			return
		}
		go handleStream(stream)
	}
}

func handleStream(stream *quic.Stream) {
	defer stream.Close()

	// 读取数据
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	msg := string(buf[:n])
	fmt.Printf("Received: %s\n", msg)

	// 回显数据
	response := fmt.Sprintf("Echo: %s", msg)
	_, err = stream.Write([]byte(response))
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
	}
}
