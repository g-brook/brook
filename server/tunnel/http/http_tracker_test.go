package http

import (
	"io"
	"testing"
	"time"

	"github.com/g-brook/brook/common/hash"
)

type fakeFuture struct{ closed bool }

func (f *fakeFuture) Done([]byte)  {}
func (f *fakeFuture) ReqId() int64 { return 1 }
func (f *fakeFuture) Close()       { f.closed = true }

func TestResponseFutureWaitTimeout(t *testing.T) {
	tracker := &Tracker{trackers: hash.NewSyncMap[int64, Future]()}
	future := newResponseFuture(tracker)
	_, err := future.WaitTimeout(10 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestTrackerCloseAllClosesTrackedFutures(t *testing.T) {
	tracker := &Tracker{trackers: hash.NewSyncMap[int64, Future]()}
	future := &fakeFuture{}
	tracker.trackers.Store(1, future)

	tracker.closeAll()

	if !future.closed {
		t.Fatal("expected future to be closed")
	}
	if _, ok := tracker.trackers.Load(1); ok {
		t.Fatal("expected future to be removed from tracker map")
	}
}

func TestWsFutureCloseMakesReadEOF(t *testing.T) {
	tracker := &Tracker{trackers: hash.NewSyncMap[int64, Future]()}
	future := newWsFuture(tracker, 100)
	future.Close()
	buf := make([]byte, 8)
	_, err := future.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}
