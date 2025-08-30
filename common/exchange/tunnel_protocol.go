package exchange

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/brook/common/aio"
	"github.com/brook/common/log"
)

var (
	v1        int8  = 1
	lenLen    int32 = 4
	reqIdLen  int32 = 8
	verLen    int32 = 1
	headerLen       = lenLen + reqIdLen + verLen
)

type TunnelProtocol struct {
	Len   int32
	ReqId int64
	Ver   int8
	Data  []byte
}

// NewTunnelWriter   creates a new instance of TunnelProtocol with the provided data
// It initializes the protocol version to v1 and increments the request ID counter
func NewTunnelWriter(data []byte, reqId int64) *TunnelProtocol {
	return &TunnelProtocol{
		Len:   headerLen + int32(len(data)),
		ReqId: reqId,
		Ver:   v1, // Set the protocol version to v1
		Data:  data,
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
	lens := make([]byte, lenLen)
	_, err := io.ReadFull(r, lens)
	if err != nil {
		return err
	}
	t.Len = int32(binary.BigEndian.Uint32(lens))
	if t.Len < headerLen {
		log.Error("packet size error")
		return errors.New("invalid packet size")
	}
	data := make([]byte, t.Len-lenLen)
	if _, err := io.ReadFull(r, data); err != nil {
		log.Error(err.Error())
		return err
	}
	t.Decode(data)
	return nil
}

func (t *TunnelProtocol) Encode() []byte {
	pool := aio.GetBufPool(int(t.Len))
	var bufData []byte
	_ = aio.WithBuf(func(buf *bytes.Buffer) error {
		err := binary.Write(buf, binary.BigEndian, t.Len)
		_ = binary.Write(buf, binary.BigEndian, t.Ver)
		_ = binary.Write(buf, binary.BigEndian, t.ReqId)
		_, _ = buf.Write(t.Data)
		bufData = buf.Bytes()
		return err
	}, pool)
	return bufData
}

func (t *TunnelProtocol) Decode(data []byte) {
	t.Ver = int8(data[0])
	t.ReqId = int64(binary.BigEndian.Uint64(data[1:9]))
	t.Data = data[9:]
}
