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
	"errors"
	"io"

	"github.com/brook/common/log"
)

var (
	V1          int8 = 1
	WebsocketV1 int8 = 2
	WebsocketV2 int8 = 3

	//protocol defined.
	lenSize    int32 = 4
	_reqIdSize int32 = 8
	verSize    int32 = 1
	attrSize   int32 = 4
	headerLen        = lenSize + _reqIdSize + verSize + attrSize
)

type TunnelProtocol struct {
	Len     int32
	ReqId   int64
	Ver     int8
	AttrLen int32
	Attr    []byte
	Data    []byte
}

// NewTunnelWriter   creates a new instance of TunnelProtocol with the provided data
// It initializes the protocol version to v1 and increments the request ID counter
func NewTunnelWriter(data []byte, reqId int64) *TunnelProtocol {
	return &TunnelProtocol{
		Len:     headerLen + int32(len(data)),
		ReqId:   reqId,
		Ver:     V1, // Set the protocol version to v1
		Data:    data,
		Attr:    []byte{},
		AttrLen: 0,
	}
}

func NewTunnelWebsocketWriterV1(data []byte, attr []byte, reqId int64) *TunnelProtocol {
	return &TunnelProtocol{
		Len:     headerLen + int32(len(attr)) + int32(len(data)),
		ReqId:   reqId,
		Ver:     WebsocketV1,
		Data:    data,
		AttrLen: int32(len(attr)),
		Attr:    attr,
	}
}

func NewTunnelWebsocketWriterV2(data []byte, attr []byte, reqId int64) *TunnelProtocol {
	return &TunnelProtocol{
		Len:     headerLen + int32(len(attr)) + int32(len(data)),
		ReqId:   reqId,
		Ver:     WebsocketV2,
		AttrLen: int32(len(attr)),
		Attr:    attr,
		Data:    data,
	}
}

func NewTunnelRead() *TunnelProtocol {
	return &TunnelProtocol{}
}

// Writer is a method of TunnelProtocol that handles writing data to the given io.Writer
// It takes a writer as parameter and returns an error if any occurs during the write operation
func (t *TunnelProtocol) Writer(w io.Writer) error {
	buf := t.Encode()
	_, err := w.Write(buf)
	return err
}

func (t *TunnelProtocol) Read(r io.Reader) error {
	lens := make([]byte, lenSize)
	_, err := io.ReadFull(r, lens)
	if err != nil {
		return err
	}
	t.Len = int32(binary.BigEndian.Uint32(lens))
	if t.Len < headerLen {
		log.Error("packet size error")
		return errors.New("invalid packet size")
	}
	data := make([]byte, t.Len-lenSize)
	if _, err := io.ReadFull(r, data); err != nil {
		log.Error(err.Error())
		return err
	}
	t.Decode(data)
	return nil
}

func (t *TunnelProtocol) Encode() []byte {
	buf := make([]byte, t.Len)
	binary.BigEndian.PutUint32(buf[0:4], uint32(t.Len))
	buf[4] = byte(t.Ver)
	binary.BigEndian.PutUint64(buf[5:13], uint64(t.ReqId))
	binary.BigEndian.PutUint32(buf[13:17], uint32(t.AttrLen))
	attrLen := 17 + t.AttrLen
	if t.AttrLen > 0 {
		copy(buf[17:attrLen], t.Attr)
	}
	copy(buf[attrLen:], t.Data)
	return buf
}

func (t *TunnelProtocol) Decode(data []byte) {
	t.Ver = int8(data[0])
	t.ReqId = int64(binary.BigEndian.Uint64(data[1:9]))
	t.AttrLen = int32(binary.BigEndian.Uint32(data[9:13]))
	dataLen := t.AttrLen + 13
	if t.AttrLen > 0 {
		t.Attr = data[13:dataLen]
	}
	t.Data = data[dataLen:]
}
