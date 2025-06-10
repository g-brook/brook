package exchange

import (
	bytes2 "bytes"
	"common/log"
	"encoding/binary"
	"io"
)

//Define protocol processing for internal data exchange.

const versionSize = 1

const totalLenSize = 4

const headerSize = totalLenSize + versionSize

const (
	Version1 int8 = 1
)

var codecByVersion map[int8]Codec

func init() {
	codecByVersion = make(map[int8]Codec)
	codecByVersion[Version1] = BobCodec{}
}

type Codec interface {
	//
	// encode
	//  @Description: encode.
	//  @param change
	//  @return []byte
	//
	encode(change ExChange) []byte

	//
	// decode
	//  @Description: decode.
	//  @param change
	//  @return ExChange
	//
	decode(bytes []byte) ExChange
}

// ExChange
// @Description: This is Data exchange protocol.
type ExChange struct {

	//from targetId.
	TargetId string

	//version.
	Version int8

	//Tunnel port
	Port int32

	//source data.
	Data []byte
}

func NewExChange(targetId string, port int32) ExChange {
	return ExChange{
		Version:  Version1,
		TargetId: targetId,
		Port:     port,
		Data:     make([]byte, 0),
	}
}

func Encode(change ExChange) []byte {
	codec, ok := codecByVersion[change.Version]
	if !ok {
		log.Warn("Not found exchange codec ")
	}
	bytes := codec.encode(change)
	if bytes == nil {
		bytes = make([]byte, 0)
	}
	totalLen := len(bytes) + headerSize
	b := new(bytes2.Buffer)
	_ = binary.Write(b, binary.BigEndian, int32(totalLen))
	_ = binary.Write(b, binary.BigEndian, change.Version)
	_ = binary.Write(b, binary.BigEndian, bytes)
	return b.Bytes()
}

func Decode(reader io.Reader) ExChange {
	var rsp = ExChange{}
	lenByte := make([]byte, totalLenSize)
	if _, err := io.ReadFull(reader, lenByte); err != nil {
		log.Error(err.Error())
		return rsp
	}
	dataLen := binary.BigEndian.Uint32(lenByte)
	if dataLen < headerSize {
		log.Error("packet size error")
		return rsp
	}
	data := make([]byte, dataLen-totalLenSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		log.Error(err.Error())
		return rsp
	}
	ver := int8(data[0])
	codec, ok := codecByVersion[ver]
	if !ok {
		log.Error("data bob version is error %d", ver)
		return rsp
	}
	return codec.decode(data[1:])
}
