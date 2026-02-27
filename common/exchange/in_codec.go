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

	"github.com/g-brook/brook/common/log"
	"github.com/panjf2000/gnet/v2"
)

const totalPacketSize = 4

const cmdSize = 1

const ptypeSize = 1

const reqIdSize = 8

const rspCode = 2

const headerSize = totalPacketSize + cmdSize + ptypeSize + reqIdSize + rspCode

var ErrNeedMoreData = errors.New("need more data")

// Encoder
//
//	@Description: encoder.
//	@param data
//	@return []byte
func Encoder(data *Protocol) []byte {
	if !data.IsSuccess() && data.PType == RESPONSE {
		bytes := []byte(data.RspMsg)
		data.Data = bytes
	}
	totalLen := len(data.Data) + headerSize

	bytes := make([]byte, totalLen)
	binary.BigEndian.PutUint32(bytes[0:4], uint32(totalLen))
	bytes[4] = byte(data.Cmd)
	bytes[5] = byte(data.PType)
	binary.BigEndian.PutUint64(bytes[6:14], uint64(data.ReqId))
	binary.BigEndian.PutUint16(bytes[14:16], uint16(data.RspCode))
	copy(bytes[16:], data.Data)
	return bytes
}

func GetByteLen(lenByte []byte) int {
	return int(binary.BigEndian.Uint32(lenByte))
}

func GetBody(bodies []byte) (*Protocol, error) {
	var req Protocol
	//cmd. 0
	req.Cmd = Cmd(bodies[0])
	//type. 1
	req.PType = PType(bodies[1])
	//reqId. 2~9
	req.ReqId = int64(binary.BigEndian.Uint64(bodies[2:10]))
	//rspCode.
	rspBytes := bodies[10:12]
	req.RspCode = RspCode(int16(binary.BigEndian.Uint16(rspBytes)))
	//data.
	req.Data = bodies[12:]
	if req.PType == RESPONSE {
		if !req.IsSuccess() {
			req.RspMsg = string(req.Data)
			req.Data = nil
		}
	}
	return &req, nil
}

func Decoder(reader io.Reader) (*Protocol, error) {
	lenByte := make([]byte, totalPacketSize)
	if _, err := io.ReadFull(reader, lenByte); err != nil {
		return nil, err
	}
	dataLen := GetByteLen(lenByte)
	if dataLen < headerSize {
		log.Error("packet size error")
		return nil, errors.New("invalid packet size")
	}
	data := make([]byte, dataLen-totalPacketSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return GetBody(data)
}

func Decoder2(rb gnet.Conn) (int, *Protocol, error) {
	if rb.InboundBuffered() < totalPacketSize {
		return 0, nil, ErrNeedMoreData
	}
	lenByte, _ := rb.Peek(totalPacketSize)
	dataLen := GetByteLen(lenByte)
	if dataLen <= 0 {
		return 0, nil, errors.New("invalid packet size")
	}
	if dataLen < headerSize {
		log.Error("packet size error")
		return 0, nil, errors.New("invalid packet size")
	}
	if rb.InboundBuffered() < dataLen {
		return 0, nil, ErrNeedMoreData
	}
	frame, _ := rb.Peek(dataLen)
	body, err := GetBody(frame[totalPacketSize:])
	return dataLen, body, err
}
