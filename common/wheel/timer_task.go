package wheel

import (
	"sync"
	"time"
)

type TaskEntity interface {
	GetTimer() Timer

	GetTimerTask() *TimerTask

	Cancelled() bool

	Cancel()
}

type TimerTaskRun func(entity TaskEntity)

type TimerTask struct {
	DelayMs int64

	timerTaskEntity *TimerTaskEntity

	mu sync.Mutex

	run TimerTaskRun
}

type TimerTaskEntity struct {
	timer Timer

	timerTask *TimerTask

	expirationsMs int64

	list *TimerTaskList

	next *TimerTaskEntity

	prev *TimerTaskEntity
}

// NewTimerTaskEntity creates and returns a new TimerTaskEntity instance
// Parameters:
//   - wheel: The Timer interface that this task belongs to
//   - expirationsMs: The duration after which the task should expire, converted to milliseconds
//   - timerTask: The TimerTask to be executed when the wheel expires
//
// Returns:
//   - A pointer to the newly created TimerTaskEntity
func NewTimerTaskEntity(timer Timer,
	expirationsMs time.Duration,
	timerTask *TimerTask) *TimerTaskEntity {
	tte := &TimerTaskEntity{timer: timer, expirationsMs: expirationsMs.Milliseconds(), timerTask: timerTask}
	timerTask.setTimerTaskEntry(tte)
	return tte
}

func (t *TimerTaskEntity) GetTimer() Timer {
	return t.timer
}

func (t *TimerTaskEntity) GetTimerTask() *TimerTask {
	return t.timerTask
}

func (t *TimerTaskEntity) Cancelled() bool {
	return t.timerTask.timerTaskEntity == t
}

func (t *TimerTaskEntity) Cancel() {
	if t.Cancelled() {
		t.timerTask.cancel()
	}
}

func (t *TimerTaskEntity) remove() {
	currentList := t.list
	for currentList != nil {
		currentList.remove(t)
		currentList = t.list
	}
}

func NewTimerTask(duration time.Duration, run TimerTaskRun) *TimerTask {
	return &TimerTask{DelayMs: duration.Milliseconds(), run: run}
}

func (t *TimerTask) setTimerTaskEntry(entry *TimerTaskEntity) {
	if t.timerTaskEntity != nil && t.timerTaskEntity != entry {
		t.timerTaskEntity.remove()
	}
	t.timerTaskEntity = entry
}

func (t *TimerTask) cancel() {
	if t.timerTaskEntity != nil {
		t.timerTaskEntity.remove()
	}
	t.timerTaskEntity = nil
}

func (t *TimerTask) IsCancelled() bool {
	return false
}
