package exchange

import (
	"context"
	"io"
)

type BytesBucketRead func(heads, bodies []byte, rw io.ReadWriteCloser)

type BytesBucket struct {
	//open channel.
	rw io.ReadWriteCloser

	bucketPush chan []byte

	bucketHandler map[any]BytesBucketRead

	cannelCtx context.Context

	cannel context.CancelFunc

	headerLength int

	witch func(bytes []byte) (any, int)
}

func NewBytesBucket(rw io.ReadWriteCloser, headerLength int, ctx context.Context) *BytesBucket {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &BytesBucket{
		headerLength:  headerLength,
		rw:            rw,
		bucketPush:    make(chan []byte, 1000),
		bucketHandler: make(map[any]BytesBucketRead),
		cannel:        cancelFunc,
		cannelCtx:     cancelCtx,
	}
}

func (m *BytesBucket) Run() {
	go m.revLoop()
	go m.readLoop()
}

func (m *BytesBucket) Push(bytes []byte) error {
	select {
	case <-m.cannelCtx.Done():
		return io.EOF
	case m.bucketPush <- bytes:
		return nil
	}
}

func (m *BytesBucket) AddHandler(cmd any, bucket BytesBucketRead) {
	m.bucketHandler[cmd] = bucket
}

func (m *BytesBucket) revLoop() {
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
			_, _ = m.rw.Write(message)
		}
	}
}

func (m *BytesBucket) readLoop() {
	readFunction := func() error {
		maybeHeader := make([]byte, m.headerLength)
		_, err := m.rw.Read(maybeHeader)
		if err != nil {
			return err
		}
		cmd, length := m.witch(maybeHeader)
		if handler, ok := m.bucketHandler[cmd]; ok {
			body := make([]byte, length-m.headerLength)
			_, err = m.rw.Read(body)
			if err != nil {
				return err
			}
			handler(maybeHeader, body, m.rw)
		}
		return nil
	}
	for {
		err := readFunction()
		if err == io.EOF {
			m.cannel()
			return
		}
	}
}

func (m *BytesBucket) Done() <-chan struct{} {
	return m.cannelCtx.Done()
}

// Close the bucket
func (m *BytesBucket) Close() {
	m.cannel()
}
