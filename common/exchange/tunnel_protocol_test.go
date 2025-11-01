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

package exchange

import (
	"encoding/binary"
	"testing"

	"github.com/brook/common/log"
)

func TestNewTunnelRead(t *testing.T) {
	log.NewLogger(nil)
	bytes := []byte("test")
	bytes2 := []byte("attr")
	writer := NewTunnelWebsocketWriterV2(bytes, bytes2, 100000)
	encode := writer.Encode()
	t.Log(encode)
	t.Log("总长度:", len(encode))
	read := NewTunnelRead()
	read.Len = int32(binary.BigEndian.Uint32(encode[:4]))
	t.Log(read.Len - lenSize)
	read.Decode(encode[4:])
	t.Log("Attr:", string(read.Attr))
	t.Log("Data:", string(read.Data))
	//assert.Equal(t, len(encode), read.Len)
	t.Log(read.Attr)
	t.Log(read.Ver)
	t.Log(read.Len)
	t.Log(read.ReqId)
	//t.Log(read)
}
