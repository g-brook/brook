// Package aio provides buffer pool implementations for efficient memory management
package aio

import (
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
	pool1k = newBufferPoll(poolSize1k)
	pool2k = newBufferPoll(poolSize2k)
	pool4k = newBufferPoll(poolSize4k)
	pool8k = newBufferPoll(poolSize8k)
	pool16k = newBufferPoll(poolSize16k)
	pool32k = newBufferPoll(poolSize32k)
}

// ByteBufPool represents a pool of byte buffers
// It wraps sync.Pool to provide a type-safe interface for []byte slices
type ByteBufPool struct {
	pool *sync.Pool
}

// Get retrieves a byte buffer from the pool
// If the pool is empty, a new buffer will be created
func (b *ByteBufPool) Get() []byte {
	byts := b.pool.Get().([]byte)
	return byts
}

// Put returns a byte buffer to the pool for reuse
func (b *ByteBufPool) Put(bytes []byte) {
	b.pool.Put(bytes)
}

// GetBuffPool32k returns a buffer pool optimized for 32KB buffers
func GetBuffPool32k() *ByteBufPool {
	return GetBufPool(poolSize32k)
}

// GetBuffPool16k returns a buffer pool optimized for 16KB buffers
func GetBuffPool16k() *ByteBufPool {
	return GetBufPool(poolSize16k)
}

// GetBuffPool8k returns a buffer pool optimized for 8KB buffers
func GetBuffPool8k() *ByteBufPool {
	return GetBufPool(poolSize8k)
}

// GetBuffPool4k returns a buffer pool optimized for 4KB buffers
func GetBuffPool4k() *ByteBufPool {
	return GetBufPool(poolSize4k)
}

// GetBuffPool2k returns a buffer pool optimized for 2KB buffers
func GetBuffPool2k() *ByteBufPool {
	return GetBufPool(poolSize2k)
}

// GetBuffPool1k returns a buffer pool optimized for 1KB buffers
func GetBuffPool1k() *ByteBufPool {
	return GetBufPool(poolSize1k)
}

// GetBufPool returns an appropriate buffer pool based on the requested size
// If no suitable pool exists, creates a new one with the specified size
func GetBufPool(size int) *ByteBufPool {
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
		return newBufferPoll(size)
	}
	return bufPool.(*ByteBufPool)
}

// newBufferPoll creates a new ByteBufPool with the specified buffer size
func newBufferPoll(size int) *ByteBufPool {
	return &ByteBufPool{
		pool: &sync.Pool{New: newPoll(size)},
	}
}

// newPoll returns a function that creates new byte slices of the specified size
// This function is used as the New field of sync.Pool
func newPoll(size int) func() interface{} {
	return func() interface{} {
		return make([]byte, size)
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
	defer bufPool.Put(buf)
	return f(buf)
}
