package queue

type MemoryQueue[T any] struct {
	data chan T
}

func NewMemoryQueue[T any](capacity int) *MemoryQueue[T] {
	return &MemoryQueue[T]{
		data: make(chan T, capacity),
	}
}

func (q *MemoryQueue[T]) Push(item T) {
	q.data <- item
}

func (q *MemoryQueue[T]) TryPush(item T) bool {
	select {
	case q.data <- item:
		return true
	default:
		return false
	}
}

func (q *MemoryQueue[T]) Pop() T {
	return <-q.data
}

func (q *MemoryQueue[T]) TryPop() (T, bool) {
	select {
	case item := <-q.data:
		return item, true
	default:
		var zero T
		return zero, false
	}
}

func (q *MemoryQueue[T]) Len() int {
	return len(q.data)
}

func (q *MemoryQueue[T]) Cap() int {
	return cap(q.data)
}
