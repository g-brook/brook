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

// Package aio provides buffer pool implementations for efficient memory management
package aio

import (
	"bytes"
	"sync"
)

// Pre-defined buffer pools of different sizes
var (
	// Buffer pools for different sizes from 1KB to 32KB
	pool1k  *ByteBufPool
	pool2k  *ByteBufPool
	pool4k  *ByteBufPool
	pool8k  *ByteBufPool
	pool16k *ByteBufPool
	pool32k *ByteBufPool

	pool1kBuf  *BufPool
	pool2kBuf  *BufPool
	pool4kBuf  *BufPool
	pool8kBuf  *BufPool
	pool16kBuf *BufPool
	pool32kBuf *BufPool

	// Pool sizes in bytes
	poolSize1k  = 1 * 1024  // 1KB
	poolSize2k  = 2 * 1024  // 2KB
	poolSize4k  = 4 * 1024  // 4KB
	poolSize8k  = 8 * 1024  // 8KB
	poolSize16k = 16 * 1024 // 16KB
	poolSize32k = 32 * 1024 // 32KB
)

// init initializes all pre-defined buffer pools with their respective sizes
func init() {
	pool1k = newByteBufferPool(poolSize1k)
	pool2k = newByteBufferPool(poolSize2k)
	pool4k = newByteBufferPool(poolSize4k)
	pool8k = newByteBufferPool(poolSize8k)
	pool16k = newByteBufferPool(poolSize16k)
	pool32k = newByteBufferPool(poolSize32k)

	pool1kBuf = newBufferPool(poolSize1k)
	pool2kBuf = newBufferPool(poolSize2k)
	pool4kBuf = newBufferPool(poolSize4k)
	pool8kBuf = newBufferPool(poolSize8k)
	pool16kBuf = newBufferPool(poolSize16k)
	pool32kBuf = newBufferPool(poolSize32k)
}

// ByteBufPool represents a pool of byte buffers
// It wraps sync.Pool to provide a type-safe interface for []byte slices
type ByteBufPool struct {
	pool *sync.Pool
}

type BufPool struct {
	pool *sync.Pool
}

func (b *ByteBufPool) Get() []byte {
	byts := b.pool.Get().([]byte)
	return byts
}

func (b *ByteBufPool) Put(bytes []byte) {
	b.pool.Put(bytes)
}

func (b *BufPool) Get() *bytes.Buffer {
	return b.pool.Get().(*bytes.Buffer)
}

func (b *BufPool) Put(buf *bytes.Buffer) {
	b.pool.Put(buf)
}

// GetBytePool32k returns a buffer pool optimized for 32KB buffers
func GetBytePool32k() *ByteBufPool {
	return GetByteBufPool(poolSize32k)
}

// GetBytePool16k returns a buffer pool optimized for 16KB buffers
func GetBytePool16k() *ByteBufPool {
	return GetByteBufPool(poolSize16k)
}

// GetBytePool8k returns a buffer pool optimized for 8KB buffers
func GetBytePool8k() *ByteBufPool {
	return GetByteBufPool(poolSize8k)
}

// GetBytePool4k returns a buffer pool optimized for 4KB buffers
func GetBytePool4k() *ByteBufPool {
	return GetByteBufPool(poolSize4k)
}

// GetBytePool2k returns a buffer pool optimized for 2KB buffers
func GetBytePool2k() *ByteBufPool {
	return GetByteBufPool(poolSize2k)
}

// GetBuffPool1k returns a buffer pool optimized for 1KB buffers
func GetBytePool1k() *ByteBufPool {
	return GetByteBufPool(poolSize1k)
}

// GetBufPool32k  returns a buffer pool optimized for 32KB buffers
func GetBufPool32k() *BufPool {
	return GetBufPool(poolSize32k)
}

// GetBufPool16k  returns a buffer pool optimized for 16KB buffers
func GetBufPool16k() *BufPool {
	return GetBufPool(poolSize16k)
}

// GetBufPool8k  returns a buffer pool optimized for 8KB buffers
func GetBufPool8k() *BufPool {
	return GetBufPool(poolSize8k)
}

// GetBufPool4k  returns a buffer pool optimized for 4KB buffers
func GetBufPool4k() *BufPool {
	return GetBufPool(poolSize4k)
}

// GetBufPool2k  returns a buffer pool optimized for 2KB buffers
func GetBufPool2k() *BufPool {
	return GetBufPool(poolSize2k)
}

// GetBufPool1k  returns a buffer pool optimized for 1KB buffers
func GetBufPool1k() *BufPool {
	return GetBufPool(poolSize1k)
}

// GetByteBufPool returns an appropriate buffer pool based on the requested size
// If no suitable pool exists, creates a new one with the specified size
func GetByteBufPool(size int) *ByteBufPool {
	var bufPool interface{}
	switch {
	case size >= poolSize32k:
		bufPool = pool32k
	case size >= poolSize16k:
		bufPool = pool16k
	case size >= poolSize8k:
		bufPool = pool8k
	case size >= poolSize4k:
		bufPool = pool4k
	case size >= poolSize2k:
		bufPool = pool2k
	case size >= poolSize1k:
		bufPool = pool1k
	}
	if bufPool == nil {
		return newByteBufferPool(size)
	}
	return bufPool.(*ByteBufPool)
}

func GetBufPool(size int) *BufPool {
	var bufPool interface{}
	switch {
	case size >= poolSize32k:
		bufPool = pool32kBuf
	case size >= poolSize16k:
		bufPool = pool16kBuf
	case size >= poolSize8k:
		bufPool = pool8kBuf
	case size >= poolSize4k:
		bufPool = pool4kBuf
	case size >= poolSize2k:
		bufPool = pool2kBuf
	case size >= poolSize1k:
		bufPool = pool1kBuf
	}
	if bufPool == nil {
		return newBufferPool(size)
	}
	return bufPool.(*BufPool)
}

// newBufferPool creates a new ByteBufPool with the specified buffer size
// newByteBufferPool creates a new byte buffer pool with specified size
// It returns a pointer to ByteBufPool structure initialized with a sync.Pool
// The pool's New function is set to newPoll with the given size parameter
func newByteBufferPool(size int) *ByteBufPool {
	return &ByteBufPool{
		// pool is a sync.Pool instance that manages byte buffer recycling
		// New function is initialized with newPoll to create new buffers when pool is empty
		pool: &sync.Pool{New: newPoll(size, true)},
	}
}

// newBufferPool creates a new buffer pool with specified size
// It initializes a sync.Pool with a custom New function that creates buffers of given size
func newBufferPool(size int) *BufPool {
	return &BufPool{
		// Initialize the pool with a custom New function
		pool: &sync.Pool{New: newPoll(size, false)},
	}
}

// newPoll returns a function that creates new byte slices of the specified size
// This function is used as the New field of sync.Pool
func newPoll(size int, isBytes bool) func() interface{} {
	return func() interface{} {
		byt := make([]byte, size)
		if isBytes {
			return byt
		}
		return bytes.NewBuffer(byt)
	}
}

// WithBuffer is a helper function that manages buffer lifecycle operations
// It gets a buffer from the pool, executes the provided function with the buffer,
// and ensures the buffer is returned to the pool afterwards
func WithBuffer(f func(buf []byte) error, bufPool *ByteBufPool) error {
	if bufPool == nil {
		panic("buf pool is null")
	}
	if f == nil {
		panic("function is null ")
	}
	buf := bufPool.Get()
	defer func() {
		bufPool.Put(buf)
	}()
	return f(buf)
}

func WithBuf(f func(buf *bytes.Buffer) error, bufPool *BufPool) error {
	if bufPool == nil {
		panic("buf pool is null")
	}
	if f == nil {
		panic("function is null ")
	}
	buf := bufPool.Get()
	buf.Reset()
	defer func() {
		bufPool.Put(buf)
	}()
	return f(buf)
}
