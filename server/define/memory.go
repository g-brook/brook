package defin

import (
	"sync"
)

type KeyType string

var Store Memory

func init() {
	Store = Memory{
		values: make(map[KeyType]interface{}),
	}
}

type Memory struct {
	values map[KeyType]any
	rn     sync.RWMutex
}

func Get[T any](key KeyType) T {
	a := Store.values[key]
	return a.(T)
}

func Set(key KeyType, value any) {
	Store.rn.Lock()
	defer Store.rn.Unlock()
	Store.values[key] = value
}

func Delete(key KeyType) {
	Store.rn.Lock()
	defer Store.rn.Unlock()
	delete(Store.values, key)
}
