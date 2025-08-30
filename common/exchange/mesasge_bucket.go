package exchange

import (
	"context"
	"io"
	"time"

	"github.com/brook/common/log"
)

// BucketRead is a function type that takes a pointer to a Protocol struct and an io.ReadWriteCloser as parameters
type BucketRead func(p *Protocol, rw io.ReadWriteCloser)

// MessageBucket is a struct that contains a BytesBucket, a channel for pushing Protocol structs, a map of Cmd to BucketRead functions, and a default BucketRead function
type MessageBucket struct {
	//open channel.
	bytesBucket *BytesBucket

	bucketPush chan *Protocol

	bucketHandler map[Cmd]BucketRead

	defaultHandler BucketRead

	defaultReader func() error
}

// NewMessageBucket creates a new MessageBucket struct and returns a pointer to it
func NewMessageBucket(rw io.ReadWriteCloser, ctx context.Context) *MessageBucket {
	return &MessageBucket{
		bytesBucket:   NewBytesBucket(rw, 4, ctx),
		bucketPush:    make(chan *Protocol, 1000),
		bucketHandler: make(map[Cmd]BucketRead),
	}
}

// SetDefaultHandler This function sets the default handler for the MessageBucket struct
func (m *MessageBucket) SetDefaultHandler(def BucketRead) {
	// Assign the passed in BucketRead function to the defaultHandler field of the MessageBucket struct
	m.defaultHandler = def
}

// read is a function that takes a pointer to a Protocol struct, a byte slice, and an io.ReadWriteCloser as parameters
func (m *MessageBucket) read(_, bytes []byte, rw io.ReadWriteCloser) {
	body, err := GetBody(bytes)
	if err != nil {
		log.Warn("error.")
		return
	}
	// Sync request complete.
	completeOk := Tracker.Complete(body)
	if completeOk {
		return
	}
	// Check if the Cmd field of the Protocol struct is in the bucketHandler map
	if handler, ok := m.bucketHandler[body.Cmd]; ok {
		handler(body, rw)
	} else {
		if m.defaultHandler != nil {
			m.defaultHandler(body, rw)
		}
	}
}

func (m *MessageBucket) SetReaderFunction(fun ReadFunction) {
	m.bytesBucket.SetReadFunction(fun)
}

// Run is a function that adds a handler to the bytesBucket and starts the bytesBucket
func (m *MessageBucket) Run() {
	m.bytesBucket.AddHandler("Protocol", m.read)
	m.bytesBucket.witch = func(bytes []byte) (any, int) {
		return "Protocol", GetByteLen(bytes)
	}
	m.bytesBucket.doRunning(func(revLoop, readLoop func()) {
		revLoop()
		readLoop()
	})
}

// Push is a function that takes a pointer to a Protocol struct and pushes it to the bucketPush channel
func (m *MessageBucket) Push(message *Protocol) error {
	return m.bytesBucket.Push(message.Bytes())
}

func (m *MessageBucket) PushWitchRequest(message InBound) error {
	request, _ := NewRequest(message)
	return m.Push(request)
}

// SyncPushWitchProtocol This function takes a pointer to a MessageBucket and a pointer to a Protocol as parameters and returns a pointer to a Protocol and an error.
func (m *MessageBucket) SyncPushWitchProtocol(message *Protocol) (*Protocol, error) {
	// This function calls the SyncWriteByProtocol function with the message, a 10 second timeout, and a function that calls the Push function on the MessageBucket.
	return SyncWriteByProtocol(message, 10*time.Second, func(p *Protocol) error {
		return m.Push(p)
	})
}

func (m *MessageBucket) SyncPushWithRequest(message InBound) (*Protocol, error) {
	return SyncWriteInBound(message, 10*time.Second, func(p *Protocol) error {
		return m.Push(p)
	})
}

// AddHandler is a function that takes a Cmd and a BucketRead function and adds them to the bucketHandler map
func (m *MessageBucket) AddHandler(cmd Cmd, bucket BucketRead) {
	if _, ok := m.bucketHandler[cmd]; !ok {
		m.bucketHandler[cmd] = bucket
	}
}

// Done is a function that returns a channel that is closed when the bytesBucket is done
func (m *MessageBucket) Done() <-chan struct{} {
	return m.bytesBucket.Done()
}

// Close the bucket
func (m *MessageBucket) Close() {
	m.bytesBucket.Close()
}
