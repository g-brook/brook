package defin

import (
	"github.com/brook/common"
	"sync"
)

var Store Memory

func init() {
	Store = Memory{
		values: make(map[common.KeyType]interface{}),
	}
}

type Memory struct {
	values map[common.KeyType]any
	rn     sync.RWMutex
}

func Get[T any](key common.KeyType) T {
	a := Store.values[key]
	return a.(T)
}

func Set(key common.KeyType, value any) {
	Store.rn.Lock()
	defer Store.rn.Unlock()
	Store.values[key] = value
}

func Delete(key common.KeyType) {
	Store.rn.Lock()
	defer Store.rn.Unlock()
	delete(Store.values, key)
}
