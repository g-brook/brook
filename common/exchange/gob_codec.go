package exchange

import (
	bytes2 "bytes"
	"encoding/gob"
)

// BobCodec
// @Description: version bob codec.
type BobCodec struct {
}

func (g BobCodec) encode(change ExChange) []byte {
	var buf bytes2.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(change)
	return buf.Bytes()
}

func (g BobCodec) decode(bytes []byte) ExChange {
	var buf = bytes2.NewBuffer(bytes)
	dec := gob.NewDecoder(buf)
	var decoded ExChange
	_ = dec.Decode(&decoded)
	return decoded
}
