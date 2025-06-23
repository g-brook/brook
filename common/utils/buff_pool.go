package utils

import "sync"

var (
	pool1k  *ByteBufPool
	pool2k  *ByteBufPool
	pool4k  *ByteBufPool
	pool8k  *ByteBufPool
	pool16k *ByteBufPool
	pool32k *ByteBufPool

	poolSize1k  = 1 * 1024
	poolSize2k  = 2 * 1024
	poolSize4k  = 4 * 1024
	poolSize8k  = 8 * 1024
	poolSize16k = 16 * 1024
	poolSize32k = 32 * 1024
)

func init() {
	pool1k = newBufferPoll(poolSize1k)
	pool2k = newBufferPoll(poolSize2k)
	pool4k = newBufferPoll(poolSize4k)
	pool8k = newBufferPoll(poolSize8k)
	pool16k = newBufferPoll(poolSize16k)
	pool32k = newBufferPoll(poolSize32k)
}

type ByteBufPool struct {
	pool sync.Pool
}

func (b *ByteBufPool) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *ByteBufPool) Put(bytes []byte) {
	b.pool.Put(bytes)
}

func GetBuffPool32k() *ByteBufPool {
	return GetBufPool(poolSize32k)
}

func GetBuffPool16k() *ByteBufPool {
	return GetBufPool(poolSize16k)
}

func GetBuffPool8k() *ByteBufPool {
	return GetBufPool(poolSize8k)
}

func GetBuffPool4k() *ByteBufPool {
	return GetBufPool(poolSize4k)
}

func GetBuffPool2k() *ByteBufPool {
	return GetBufPool(poolSize2k)
}

func GetBuffPool1k() *ByteBufPool {
	return GetBufPool(poolSize1k)
}

func GetBufPool(size int) *ByteBufPool {
	var bufPool interface{}
	switch {
	case size > poolSize32k:
		bufPool = pool32k
	case size > poolSize16k:
		bufPool = pool16k
	case size > poolSize8k:
		bufPool = pool8k
	case size > poolSize4k:
		bufPool = pool4k
	case size > poolSize2k:
		bufPool = pool2k
	case size > poolSize1k:
		bufPool = pool1k
	}
	if bufPool == nil {
		return newBufferPoll(size)
	}
	return bufPool.(*ByteBufPool)
}

func newBufferPoll(size int) *ByteBufPool {
	return &ByteBufPool{
		pool: sync.Pool{New: newPoll(size)},
	}
}

func newPoll(size int) func() interface{} {
	return func() interface{} {
		return make([]byte, size)
	}
}
