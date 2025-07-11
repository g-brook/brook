package exchange

import (
	"context"
	"github.com/brook/common/log"
	"io"
)

type BucketRead func(p *Protocol, rw io.ReadWriteCloser)

type MessageBucket struct {
	//open channel.
	bytesBucket *BytesBucket

	bucketPush chan *Protocol

	bucketHandler map[Cmd]BucketRead
}

func NewMessageBucket(rw io.ReadWriteCloser, ctx context.Context) *MessageBucket {
	return &MessageBucket{
		bytesBucket:   NewBytesBucket(rw, 4, ctx),
		bucketPush:    make(chan *Protocol, 1000),
		bucketHandler: make(map[Cmd]BucketRead),
	}
}

func (m *MessageBucket) read(_, bytes []byte, rw io.ReadWriteCloser) {
	body, err := GetBody(bytes)
	if err != nil {
		log.Warn("error.")
		return
	}
	if handler, ok := m.bucketHandler[body.Cmd]; ok {
		handler(body, rw)
	}
}

func (m *MessageBucket) Run() {
	m.bytesBucket.AddHandler("Protocol", m.read)
	m.bytesBucket.witch = func(bytes []byte) (any, int) {
		return "Protocol", GetByteLen(bytes)
	}
	m.bytesBucket.Run()
}

func (m *MessageBucket) Push(message *Protocol) error {
	return m.bytesBucket.Push(message.Bytes())
}

func (m *MessageBucket) AddHandler(cmd Cmd, bucket BucketRead) {
	m.bucketHandler[cmd] = bucket
}

func (m *MessageBucket) Done() <-chan struct{} {
	return m.bytesBucket.Done()
}

// Close the bucket
func (m *MessageBucket) Close() {
	m.bytesBucket.Close()
}
