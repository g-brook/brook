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
	"context"
	"io"
)

type BytesBucketRead func(heads, bodies []byte, rw io.ReadWriteCloser)

type ReadFunction func(sch io.ReadWriteCloser) error

type BytesBucket struct {
	//open channel.
	rw io.ReadWriteCloser

	bucketPush chan []byte

	bucketHandler map[any]BytesBucketRead

	cannelCtx context.Context

	cannel context.CancelFunc

	headerLength int

	witch func(bytes []byte) (any, int)

	readFunction ReadFunction
}

func NewBytesBucket(rw io.ReadWriteCloser,
	headerLength int,
	ctx context.Context) *BytesBucket {
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

func (m *BytesBucket) doRunning(fun func(revLoop, readLoop func())) {
	fun(func() {
		go m.revLoop()
	}, func() {
		go m.readLoop()
	})
}

func (m *BytesBucket) Push(bytes []byte) error {
	select {
	case <-m.cannelCtx.Done():
		return io.EOF
	case m.bucketPush <- bytes:
		return nil
	}
}

// AddHandler This function adds a handler to the BytesBucket struct
func (m *BytesBucket) AddHandler(cmd any, bucket BytesBucketRead) {
	// Assign the bucket parameter to the bucketHandler map with the cmd parameter as the key
	m.bucketHandler[cmd] = bucket
}

// SetReadFunction This function sets the read function for the BytesBucket struct
func (m *BytesBucket) SetReadFunction(fun ReadFunction) {
	// Assign the passed function to the readFunction field of the BytesBucket struct
	m.readFunction = fun
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
	defaultFunction := func() error {
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
		var err error
		if m.readFunction != nil {
			err = m.readFunction(m.rw)
		} else {
			err = defaultFunction()
		}
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
