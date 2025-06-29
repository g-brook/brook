package exchange

import (
	"context"
	"io"
)

type BucketRead func(p *Protocol, rw io.ReadWriteCloser)

type MessageBucket struct {
	//open channel.
	rw io.ReadWriteCloser

	bucketPush chan *Protocol

	bucketHandler map[Cmd]BucketRead

	cannelCtx context.Context

	cannel context.CancelFunc
}

func NewMessageBucket(rw io.ReadWriteCloser, ctx context.Context) *MessageBucket {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &MessageBucket{
		rw:            rw,
		bucketPush:    make(chan *Protocol, 1000),
		bucketHandler: make(map[Cmd]BucketRead),
		cannel:        cancelFunc,
		cannelCtx:     cancelCtx,
	}
}

func (m *MessageBucket) Run() {
	go m.revLoop()
	go m.readLoop()
}

func (m *MessageBucket) Push(message *Protocol) error {
	select {
	case <-m.cannelCtx.Done():
		return io.EOF
	case m.bucketPush <- message:
		return nil
	}
}

func (m *MessageBucket) AddHandler(cmd Cmd, bucket BucketRead) {
	m.bucketHandler[cmd] = bucket
}

func (m *MessageBucket) revLoop() {
	closeFunc := func() {
		if m.rw != nil {
			m.rw.Close()
		}
	}
	for {
		select {
		case <-m.cannelCtx.Done():
			closeFunc()
			return
		case message := <-m.bucketPush:
			_, _ = m.rw.Write(message.Bytes())
		}
	}
}

func (m *MessageBucket) readLoop() {
	readFunction := func() error {
		decoder, err := Decoder(m.rw)
		if err != nil {
			return err
		}
		if handler, ok := m.bucketHandler[decoder.Cmd]; ok {
			handler(decoder, m.rw)
		}
		return nil
	}
	for {
		err := readFunction()
		if err != nil {
			m.cannel()
		}
	}
}

func (m *MessageBucket) Done() <-chan struct{} {
	return m.cannelCtx.Done()
}
