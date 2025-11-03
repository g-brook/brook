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

// file: ws_echo.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 生产请检查 Origin！
	CheckOrigin: func(r *http.Request) bool { return true },
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ws echo handler")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer conn.Close()
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			log.Println(string(msg))
			if err != nil {
				fmt.Println("read:", err)
				break
			}

		}
	}()
	for {
		// 直接原样发送回去
		if err := conn.WriteMessage(1, []byte("PONG:"+time.Now().Format("2006-01-02 15:04:05"))); err != nil {
			fmt.Println("write:", err)
			break
		}
		<-time.After(2. * time.Second)
	}
}

func echoHandler2(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// 处理错误
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	_, err = fmt.Fprintf(w, "OK:%v", now)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/ws", echoHandler)
	http.HandleFunc("/proxy1", echoHandler2)
	log.Println("listening :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
