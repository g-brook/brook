package wheel

import (
	"sync"
)

type TimingWheel struct {
	tickMs        int64
	wheelSize     int64
	taskCounter   int64
	delayQueue    *DelayQueue
	interval      int64
	buckets       []*TimerTaskList
	currentTime   int64
	overflowWheel *TimingWheel

	mu sync.Mutex
}

func NewTimingWheel(tickMs, wheelSize, startMs int64) *TimingWheel {
	return &TimingWheel{
		tickMs:        tickMs,
		wheelSize:     wheelSize,
		delayQueue:    NewDelayQueue(100),
		interval:      tickMs * wheelSize,
		buckets:       make([]*TimerTaskList, wheelSize),
		currentTime:   startMs - (startMs % tickMs),
		overflowWheel: nil,
	}
}

func (w *TimingWheel) addOverflowWheel() {
	if w.overflowWheel == nil {
		w.overflowWheel = NewTimingWheel(w.interval, w.wheelSize, w.currentTime)
	}
}

func (w *TimingWheel) add(entry *TimerTaskEntity) bool {
	ms := entry.expirationsMs
	if entry.Cancelled() {
		return false
	}

	if ms < w.currentTime+w.tickMs {
		return false
	}
	if ms < w.currentTime+w.interval {
		return w.addCurrentOrNextWheel(entry)
	}
	if w.overflowWheel == nil {
		w.addOverflowWheel()
	}
	return w.overflowWheel.add(entry)
}

func (w *TimingWheel) addCurrentOrNextWheel(entry *TimerTaskEntity) bool {
	ms := entry.expirationsMs
	vid := ms / w.tickMs
	index := vid % w.wheelSize
	bucket := w.getBucket(index)
	bucket.add(entry)
	if bucket.SetExpiration(vid * w.tickMs) {
		w.delayQueue.Offer(bucket, 1)
	}
	return true
}

func (w *TimingWheel) getBucket(index int64) *TimerTaskList {
	list := w.buckets[index]
	if list == nil {
		w.mu.Lock()
		list = w.buckets[index]
		if list == nil {
			list = NewTimerTaskList(w.taskCounter)
			w.buckets[index] = list
		}
	}
	return list
}
