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
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	ctx := context.Background()
	header := ws.HandshakeHeaderHTTP{
		"Origin": {"http://127.0.0.1"},
	}
	dialer := ws.Dialer{
		Header: header,
	}
	conn, _, _, err := dialer.Dial(ctx, "wss://127.0.0.1:30003/ws")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		// send
		_ = wsutil.WriteClientText(conn, []byte("PING"))
		// receive
		msg, err := wsutil.ReadServerText(conn)
		if err != nil {
			println(err.Error())
		} else {
			println(string(msg))
		}
		<-time.After(2 * time.Second)
	}
}
