package exchange

import (
	bytes2 "bytes"
	"encoding/binary"
	"errors"
	"github.com/brook/common/log"
	"io"
)

const totalPacketSize = 4

const cmdSize = 1

const ptypeSize = 1

const reqIdSize = 8

const rspCode = 2

const headerSize = totalPacketSize + cmdSize + ptypeSize + reqIdSize + rspCode

// Encoder
//
//	@Description: encoder.
//	@param data
//	@return []byte
func Encoder(data *Protocol) []byte {
	b := new(bytes2.Buffer)
	totalLen := len(data.Data) + headerSize
	_ = binary.Write(b, binary.BigEndian, int32(totalLen))
	_ = binary.Write(b, binary.BigEndian, int8(data.Cmd))
	_ = binary.Write(b, binary.BigEndian, int8(data.PType))
	_ = binary.Write(b, binary.BigEndian, data.ReqId)
	_ = binary.Write(b, binary.BigEndian, data.RspCode)
	_ = binary.Write(b, binary.BigEndian, data.Data)
	return b.Bytes()
}

// Decoder
//
//	@Description: decoder.
//	@param bytes
//	@param reader
//	@return inter.Protocol
//	@return error
func Decoder(reader io.Reader) (*Protocol, error) {
	lenByte := make([]byte, totalPacketSize)
	if _, err := io.ReadFull(reader, lenByte); err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(lenByte)
	if dataLen < headerSize {
		log.Error("packet size error")
		return nil, errors.New("invalid packet size")
	}
	data := make([]byte, dataLen-totalPacketSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	var req Protocol
	//cmd. 0
	req.Cmd = Cmd(data[0])
	//type. 1
	req.PType = PType(data[1])
	//reqId. 2~9
	req.ReqId = int64(binary.BigEndian.Uint64(data[2:10]))
	//rspCode.
	rspBytes := data[10:12]
	req.RspCode = RspCode(int16(binary.BigEndian.Uint16(rspBytes)))
	//data.
	req.Data = data[12:]
	return &req, nil
}
