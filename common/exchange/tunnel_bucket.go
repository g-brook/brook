package exchange

import (
	"context"
	"encoding/binary"
	"io"
	"sync/atomic"

	"github.com/brook/common/hash"
)

type TunnelBucketRead func(p *TunnelProtocol)

type TunnelBucket struct {
	bytesBucket *BytesBucket
	reqIdIndex  atomic.Int64
	requests    *hash.SyncMap[int64, TunnelBucketRead]
	defaultRead TunnelBucketRead
}

func (t *TunnelBucket) DefaultRead(defaultRead TunnelBucketRead) {
	t.defaultRead = defaultRead
}

// NewTunnelBucket creates a new TunnelBucket instance with a given ReadWriteCloser and context
// It initializes the TunnelBucket with a BytesBucket that uses the provided ReadWriteCloser,
// a buffer size of 4, and the given context
func NewTunnelBucket(rw io.ReadWriteCloser,
	ctx context.Context) *TunnelBucket {
	bucket := &TunnelBucket{
		bytesBucket: NewBytesBucket(rw, 4, ctx), // Initialize bytesBucket with a new BytesBucket instance
		requests:    hash.NewSyncMap[int64, TunnelBucketRead](),
	}
	bucket.Run()
	return bucket
}

func (t *TunnelBucket) Push(data []byte, read TunnelBucketRead) error {
	writer := NewTunnelWriter(data, t.reqIdIndex.Add(1))
	if read != nil {
		t.requests.Store(writer.ReqId, read)
	}
	return t.bytesBucket.Push(writer.Encode())
}

// Run is a method of TunnelBucket that starts the tunnel's operation
func (t *TunnelBucket) Run() {
	t.bytesBucket.AddHandler("Tunnel", t.read)
	t.bytesBucket.witch = func(bytes []byte) (any, int) {
		return "Tunnel", int(binary.BigEndian.Uint32(bytes))
	}
	t.bytesBucket.doRunning(func(revLoop, readLoop func()) {
		revLoop()
		readLoop()
	})
}

// read is a method of TunnelBucket that processes incoming bytes through a tunnel read operation
func (m *TunnelBucket) read(_, bytes []byte, _ io.ReadWriteCloser) {
	tp := NewTunnelRead()
	tp.Decode(bytes)
	var ok bool
	var load TunnelBucketRead
	if tp.ReqId != 0 {
		load, ok = m.requests.Load(tp.ReqId)
		defer m.requests.Delete(tp.ReqId)
	}
	if ok {
		load(tp)
	} else {
		if m.defaultRead != nil {
			m.defaultRead(tp)
		}
	}
}
