package hash

import "sync"

type SyncMap[K comparable, V any] struct {
	data sync.Map
}

// NewSyncMap This function creates a new SyncMap with the given key and value types
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	// Create a new SyncMap with the given key and value types
	return &SyncMap[K, V]{
		// Initialize the maps field with a new map
		//data: sync.Map{},
	}
}

// LoadOrStore This function takes a receiver of type SyncMap[K, V], a key of type K, and a value of type V, and returns a value of type V and a boolean.
// It uses the LoadOrStore method of the receiver's data field to either load the value associated with the key or store the value if it doesn't exist.
// The function then returns the actual value and a boolean indicating whether the value was loaded or stored.
func (receiver *SyncMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	// LoadOrStore method of the receiver's data field is called with the key and value
	actual, loaded := receiver.data.LoadOrStore(key, value)
	// The actual value is returned as a value of type V and the loaded boolean is returned
	return actual.(V), loaded
}

// LoadAndDelete This function takes a key of type any and returns a value of type V and a boolean indicating whether the key was found in the map.
func (receiver *SyncMap[K, V]) LoadAndDelete(key any) (V, bool) {
	// LoadAndDelete is a method of the data field of the receiver, which is of type sync.Map.
	value, loaded := receiver.data.LoadAndDelete(key)
	// If the key was not found in the map, return a zero value of type V and false.
	if !loaded {
		var zero V
		return zero, false
	}
	// If the key was found in the map, return the value and true.
	return value.(V), true
}

// Delete This function deletes a key from the SyncMap
func (receiver *SyncMap[K, V]) Delete(key K) {
	// Call the Delete function from the data field of the SyncMap
	receiver.data.Delete(key)
}

// Swap This function swaps the value associated with a given key in a SyncMap
func (receiver *SyncMap[K, V]) Swap(key K, value V) (V, bool) {
	// Swap the value in the data map
	previous, loaded := receiver.data.Swap(key, value)
	// Return the previous value and a boolean indicating whether the value was loaded
	return previous.(V), loaded
}

// CompareAndSwap This function compares the value of a key in a SyncMap with an old value and swaps it with a new value if they are equal.
func (receiver *SyncMap[K, V]) CompareAndSwap(key K, old V, new V) (swapped bool) {
	// Call the CompareAndSwap function on the receiver's data with the given key, old value, and new value
	return receiver.data.CompareAndSwap(key, old, new)
}

// CompareAndDelete This function compares the value of a key in a SyncMap with an old value and deletes the key if they match
func (receiver *SyncMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	// Call the CompareAndDelete function on the receiver's data with the key and old value
	return receiver.data.CompareAndDelete(key, old)
}

// Range This function takes a function as an argument and iterates over the SyncMap, calling the function for each key-value pair.
func (receiver *SyncMap[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	// This function takes a key and value as arguments and calls the provided function with the key and value.
	receiver.data.Range(func(key, value any) bool {
		// This line calls the provided function with the key and value and returns the result.
		return f(key.(K), value.(V))
	})
}

// Clear This function clears the data in the SyncMap
func (receiver *SyncMap[K, V]) Clear() {
	// Call the Clear function on the data field of the receiver
	receiver.data.Clear()
}

// Store This function stores a key-value pair in the SyncMap
func (receiver *SyncMap[K, V]) Store(key K, value V) {
	// Store the key-value pair in the data field of the SyncMap
	receiver.data.Store(key, value)
}

// Load This function is a method of the SyncMap struct and is used to load a value from the map given a key.
// It returns the value and a boolean indicating whether the key was found in the map.
func (receiver *SyncMap[K, V]) Load(key K) (V, bool) {
	// Load the value from the map using the key
	value, ok := receiver.data.Load(key)
	// If the key is not found, return a zero value and false
	if !ok {
		var zero V
		return zero, false
	}
	// Otherwise, return the value and true
	return value.(V), true
}
