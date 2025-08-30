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

// NewTimingWheel creates a new timing wheel with specified parameters
//
// @param tickMs: the time interval of each tick in milliseconds
// @param wheelSize: the number of buckets in the timing wheel
// @param startMs: the start time in milliseconds
// @return: a new instance of TimingWheel
func NewTimingWheel(tickMs, wheelSize, startMs int64) *TimingWheel {
	return &TimingWheel{
		tickMs:        tickMs,                            // Duration of each tick in milliseconds
		wheelSize:     wheelSize,                         // Number of buckets in the wheel
		delayQueue:    NewDelayQueue(100),                // Delay queue for managing tasks
		interval:      tickMs * wheelSize,                // Total interval of the timing wheel
		buckets:       make([]*TimerTaskList, wheelSize), // Initialize buckets for storing tasks
		currentTime:   startMs - (startMs % tickMs),      // Align start time to tick boundary
		overflowWheel: nil,                               // No overflow wheel initially
	}
}

// addOverflowWheel adds a new overflow wheel to the current timing wheel if it doesn't already exist.
// This method is used to implement hierarchical timing wheels for more efficient time management.
func (w *TimingWheel) addOverflowWheel() {
	// Check if the overflow wheel is nil (not initialized yet)
	if w.overflowWheel == nil {
		// Create a new overflow wheel with the same interval and wheel size as the parent wheel,
		// but starting at the current time of the parent wheel
		w.overflowWheel = NewTimingWheel(w.interval, w.wheelSize, w.currentTime)
	}
}

// add adds a TimerTaskEntity to the timing wheel
// @param entry the TimerTaskEntity to be added
// @return bool true if the entry was successfully added, false otherwise
func (w *TimingWheel) add(entry *TimerTaskEntity) bool {
	// Get the expiration time in milliseconds of the entry
	ms := entry.expirationsMs
	// If the entry has already been cancelled, don't add it
	if entry.Cancelled() {
		return false
	}

	// If the expiration time is earlier than the current time plus one tick, don't add it
	if ms < w.currentTime+w.tickMs {
		return false
	}
	// If the expiration time is within the current interval, add to current or next wheel
	if ms < w.currentTime+w.interval {
		return w.addCurrentOrNextWheel(entry)
	}
	// If there's no overflow wheel, create one
	if w.overflowWheel == nil {
		w.addOverflowWheel()
	}
	// Otherwise, add to the overflow wheel
	return w.overflowWheel.add(entry)
}

// addCurrentOrNextWheel adds a timer task entity to the current timing wheel or the next one if needed
// @param entry: The TimerTaskEntity to be added to the timing wheel
// @return: Returns true if the entry was successfully added
func (w *TimingWheel) addCurrentOrNextWheel(entry *TimerTaskEntity) bool {
	// Calculate the virtual time ID based on expiration milliseconds and tick duration
	ms := entry.expirationsMs
	// Determine the virtual time ID by dividing expiration milliseconds by tick duration
	vid := ms / w.tickMs
	// Calculate the bucket index in the current wheel using modulo operation
	index := vid % w.wheelSize
	// Get the bucket at the calculated index
	bucket := w.getBucket(index)
	// Add the entry to the bucket
	bucket.add(entry)
	// Set the expiration time for the bucket and add to delay queue if needed
	if bucket.SetExpiration(vid * w.tickMs) {
		w.delayQueue.Offer(bucket, 1)
	}
	// Return true indicating successful addition
	return true
}

// getBucket retrieves a TimerTaskList from the timing wheel at the specified index.
// If the bucket doesn't exist, it creates a new TimerTaskList and stores it at the index.
// This method uses double-checked locking pattern to minimize lock contention.
//
// Parameters:
//
//	index: The index of the bucket to retrieve
//
// Returns:
//
//	*TimerTaskList: The TimerTaskList at the specified index, or a newly created one if it didn't exist
func (w *TimingWheel) getBucket(index int64) *TimerTaskList {
	// Try to get the bucket without acquiring the lock first
	list := w.buckets[index]
	if list == nil {
		// If the bucket is nil, acquire the lock to check again and create if necessary
		w.mu.Lock()
		// Double-check: another goroutine might have created the bucket while we were waiting for the lock
		list = w.buckets[index]
		if list == nil {
			// Create a new TimerTaskList with the current task counter value
			list = NewTimerTaskList(w.taskCounter)
			// Store the new bucket at the specified index
			w.buckets[index] = list
		}
		// Release the lock
	}
	return list
}
