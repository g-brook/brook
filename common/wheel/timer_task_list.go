package wheel

import (
	"sync"
)

type Timer interface {
	Add(task TimerTask)
	AdvanceClock(timeoutMs int64)
	Size() int
	Shutdown()
}

type TimerTaskList struct {
	root        *TimerTaskEntity
	mu          sync.Mutex
	expiration  int64
	taskCounter int64
}

func (l *TimerTaskList) SetExpiration(expiration int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expiration = expiration
}

func NewTimerTaskList(taskCounter int64) *TimerTaskList {
	ttl := &TimerTaskList{taskCounter: taskCounter}
	root := NewTimerTaskEntity(nil, -1, nil)
	root.next = root
	root.prev = root
	return ttl
}

func (l *TimerTaskList) remove(t *TimerTaskEntity) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if t.list == l {
		t.next.prev = t.prev
		t.prev.next = t.next
		t.next = nil
		t.prev = nil
		t.list = nil
		l.taskCounter--
	}
}

func (l *TimerTaskList) flush(consumer func(entity *TimerTaskEntity)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	head := l.root.next
	for head != l.root {
		next := head.next
		l.remove(head)
		consumer(head)
		head = next
	}
	l.expiration = -1
}

func (l *TimerTaskList) add(t *TimerTaskEntity) {
	done := false
	for !done {
		t.remove()
		l.mu.Lock()
		if t.list == nil {
			tail := l.root.prev
			t.next = l.root
			t.prev = tail
			t.list = l
			tail.next = t
			l.root.prev = t
			l.taskCounter++
			done = true
		}
		l.mu.Unlock()
	}
}

func (l *TimerTaskList) Foreach(consumer func(task *TimerTask)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for entry := l.root.next; entry != l.root; {
		next := entry.next
		if !entry.timerTask.IsCancelled() {
			consumer(entry.timerTask)
		}
		entry = next
	}
}
