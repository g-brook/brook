package wheel

import (
	"testing"
	"time"
)

func TestDelayQueue_Offer(t *testing.T) {
	dq := NewDelayQueue(10)

	now := time.Now().UnixMilli()

	dq.Offer("first", now+10)
	dq.Offer("second", now+50)
	dq.Offer("third", now+30)

	exitC := make(chan struct{})
	defer close(exitC)

	go dq.Poll(exitC, func() int64 {
		return time.Now().UnixMilli()
	})

	results := []string{}
	for i := 0; i < 3; i++ {
		select {
		case v := <-dq.C:
			results = append(results, v.(string))
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for element %d", i)
		}
	}

	expectedOrder := []string{"first", "third", "second"}
	for i, v := range expectedOrder {
		if results[i] != v {
			t.Errorf("expected %s at position %d, got %s", v, i, results[i])
		}
	}
}

func TestDelayQueue_Concurrent(t *testing.T) {
	dq := NewDelayQueue(50)
	exitC := make(chan struct{})
	defer close(exitC)

	go dq.Poll(exitC, func() int64 {
		return time.Now().UnixMilli()
	})

	const n = 20
	now := time.Now().UnixMilli()

	for i := 0; i < n; i++ {
		go func(i int) {
			dq.Offer(i, now+int64(i*5))
		}(i)
	}

	received := make(map[int]bool)
	timeout := time.After(1 * time.Second)

	for i := 0; i < n; i++ {
		select {
		case v := <-dq.C:
			received[v.(int)] = true
		case <-timeout:
			t.Fatalf("timeout waiting for element %d", i)
		}
	}

	for i := 0; i < n; i++ {
		if !received[i] {
			t.Errorf("missing element %d", i)
		}
	}
}

func TestDelayQueue_EmptyTimeout(t *testing.T) {
	dq := NewDelayQueue(10)
	exitC := make(chan struct{})
	defer close(exitC)

	start := time.Now()
	go dq.Poll(exitC, func() int64 {
		return time.Now().UnixMilli()
	})

	// 队列为空，不提供元素，确保 Poll 能够超时退出
	time.Sleep(50 * time.Millisecond)

	if time.Since(start) > 1*time.Second {
		t.Errorf("DelayQueue did not handle empty correctly")
	}
}
