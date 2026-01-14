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
	"math/big"
	"time"

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
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
}
func main() {
	addr, err := quic.DialAddr(context.Background(), "127.0.0.1:7000", generateTLSConfig(), nil)
	if err != nil {
		panic(err)
	}
	for {
		stream, err := addr.OpenStream()
		if err != nil {
			panic(err)
		}
		stream.Write([]byte("hello:" + time.Now().String()))
		time.Sleep(1 * time.Second)
	}
}
