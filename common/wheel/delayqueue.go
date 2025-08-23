package wheel

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

type item struct {
	priority int64
	data     interface{}
	index    int
}

type priorityQueue []*item

// newPriorityQueue creates and returns a new priority queue with the specified capacity
// It initializes a priorityQueue slice with zero length and the given capacity
//
// Parameters:
//
//	capacity - the maximum number of elements the queue can hold before needing to resize
//
// Returns:
//
//	priorityQueue - a new priority queue slice initialized with the given capacity
func newPriorityQueue(capacity int) priorityQueue {
	return make(priorityQueue, 0, capacity)
}

// Len returns the number of elements in the priority queue
func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(data any) {

	item := data.(*item)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[:n-1]
	return item
}

func (pq *priorityQueue) PeekAndShift(now int64) (*item, int64) {
	if pq.Len() == 0 {
		return nil, 0
	}
	item := (*pq)[0]
	if item.priority > now {
		return nil, item.priority - now
	}
	heap.Remove(pq, 0)
	return item, 0
}

type DelayQueue struct {
	C        chan interface{}
	mu       sync.Mutex
	pq       priorityQueue
	sleeping int32
	wakeupC  chan struct{}
}

// NewDelayQueue creates and returns a new DelayQueue with the specified size
// It initializes the necessary channels and priority queue for the DelayQueue
func NewDelayQueue(size int) *DelayQueue {
	return &DelayQueue{
		C:        make(chan interface{}, size), // C is the main channel for passing items
		pq:       newPriorityQueue(size),       // pq is the priority queue that will store the delayed items
		wakeupC:  make(chan struct{}, 1),       // wakeupC is used for signaling when the queue needs to wake up
		sleeping: 0,                            // sleeping tracks the number of sleeping workers
	}
}

func (dq *DelayQueue) Offer(ele interface{}, expiration int64) {
	tm := &item{priority: expiration, data: ele}
	dq.mu.Lock()
	heap.Push(&dq.pq, tm)
	isHead := tm.index == 0
	dq.mu.Unlock()
	if isHead && atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
		dq.wakeupC <- struct{}{}
	}
}

func (dq *DelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		now := nowF()
		dq.mu.Lock()
		item, delta := dq.pq.PeekAndShift(now)
		dq.mu.Unlock()

		if item == nil {
			if delta == 0 {
				atomic.StoreInt32(&dq.sleeping, 1)
				select {
				case <-dq.wakeupC:
					atomic.StoreInt32(&dq.sleeping, 0)
					continue
				case <-exitC:
					return
				}
			} else {
				timer.Reset(time.Duration(delta) * time.Millisecond)
				atomic.StoreInt32(&dq.sleeping, 1)
				select {
				case <-dq.wakeupC:
					atomic.StoreInt32(&dq.sleeping, 0)
					if !timer.Stop() {
						<-timer.C
					}
					continue
				case <-timer.C:
					atomic.StoreInt32(&dq.sleeping, 1)
					continue
				case <-exitC:
					return

				}
			}
		}
		select {
		case dq.C <- item.data:
		case <-exitC:
			return
		}
	}
}
