/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wheel

import (
	"container/list"
	"errors"
	"fmt"
	"time"

	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	"github.com/brook/common/threading"
)

const drainWorkers = 8

var (
	ErrClosed   = errors.New("TimingWheel is closed already")
	ErrArgument = errors.New("incorrect task argument")
)

type (
	// Execute defines the method to execute the task.
	Execute func(key, value any)

	// A TimingWheel is a timing wheel object to schedule tasks.
	TimingWheel struct {
		interval      time.Duration
		ticker        Ticker
		slots         []*list.List
		timers        *hash.SyncMap[any, *positionEntry]
		tickedPos     int
		numSlots      int
		execute       Execute
		setChannel    chan timingEntry
		moveChannel   chan baseEntry
		removeChannel chan any
		drainChannel  chan func(key, value any)
		stopChannel   chan lang.PlaceholderType
	}

	timingEntry struct {
		baseEntry
		value   any
		circle  int
		diff    int
		removed bool
	}

	baseEntry struct {
		delay time.Duration
		key   any
	}

	positionEntry struct {
		pos  int
		item *timingEntry
	}

	timingTask struct {
		key   any
		value any
	}
)

// NewTimingWheel returns a TimingWheel.
func NewTimingWheel(interval time.Duration, numSlots int, execute Execute) (*TimingWheel, error) {
	if interval <= 0 || numSlots <= 0 || execute == nil {
		return nil, fmt.Errorf("interval: %v, slots: %d, execute: %p",
			interval, numSlots, execute)
	}

	return NewTimingWheelWithTicker(interval, numSlots, execute, NewTicker(interval))
}

// NewTimingWheelWithTicker returns a TimingWheel with the given ticker.
func NewTimingWheelWithTicker(interval time.Duration, numSlots int, execute Execute,
	ticker Ticker) (*TimingWheel, error) {
	tw := &TimingWheel{
		interval:      interval,
		ticker:        ticker,
		slots:         make([]*list.List, numSlots),
		timers:        hash.NewSyncMap[any, *positionEntry](),
		tickedPos:     numSlots - 1, // at previous virtual circle
		execute:       execute,
		numSlots:      numSlots,
		setChannel:    make(chan timingEntry),
		moveChannel:   make(chan baseEntry),
		removeChannel: make(chan any),
		drainChannel:  make(chan func(key, value any)),
		stopChannel:   make(chan lang.PlaceholderType),
	}

	tw.initSlots()
	go tw.run()

	return tw, nil
}

// Drain drains all items and executes them.
func (tw *TimingWheel) Drain(fn func(key, value any)) error {
	select {
	case tw.drainChannel <- fn:
		return nil
	case <-tw.stopChannel:
		return ErrClosed
	}
}

// MoveTimer moves the task with the given key to the given delay.
func (tw *TimingWheel) MoveTimer(key any, delay time.Duration) error {
	if delay <= 0 || key == nil {
		return ErrArgument
	}

	select {
	case tw.moveChannel <- baseEntry{
		delay: delay,
		key:   key,
	}:
		return nil
	case <-tw.stopChannel:
		return ErrClosed
	}
}

// RemoveTimer removes the task with the given key.
func (tw *TimingWheel) RemoveTimer(key any) error {
	if key == nil {
		return ErrArgument
	}

	select {
	case tw.removeChannel <- key:
		return nil
	case <-tw.stopChannel:
		return ErrClosed
	}
}

// SetTimer sets the task value with the given key to the delay.
func (tw *TimingWheel) SetTimer(key, value any, delay time.Duration) error {
	if delay <= 0 || key == nil {
		return ErrArgument
	}

	select {
	case tw.setChannel <- timingEntry{
		baseEntry: baseEntry{
			delay: delay,
			key:   key,
		},
		value: value,
	}:
		return nil
	case <-tw.stopChannel:
		return ErrClosed
	}
}

// Stop stops tw. No more actions after stopping a TimingWheel.
func (tw *TimingWheel) Stop() {
	close(tw.stopChannel)
}

func (tw *TimingWheel) drainAll(fn func(key, value any)) {
	runner := threading.NewTaskRunner(drainWorkers)
	for _, slot := range tw.slots {
		for e := slot.Front(); e != nil; {
			task := e.Value.(*timingEntry)
			next := e.Next()
			slot.Remove(e)
			e = next
			if !task.removed {
				runner.Schedule(func() {
					fn(task.key, task.value)
				})
			}
		}
	}
}

func (tw *TimingWheel) getPositionAndCircle(d time.Duration) (pos, circle int) {
	steps := int(d / tw.interval)
	pos = (tw.tickedPos + steps) % tw.numSlots
	circle = (steps - 1) / tw.numSlots

	return
}

func (tw *TimingWheel) initSlots() {
	for i := 0; i < tw.numSlots; i++ {
		tw.slots[i] = list.New()
	}
}

func (tw *TimingWheel) moveTask(task baseEntry) {
	val, ok := tw.timers.Load(task.key)
	if !ok {
		return
	}

	timer := val
	if task.delay < tw.interval {
		threading.GoSafe(func() {
			tw.execute(timer.item.key, timer.item.value)
		})
		return
	}

	pos, circle := tw.getPositionAndCircle(task.delay)
	if pos >= timer.pos {
		timer.item.circle = circle
		timer.item.diff = pos - timer.pos
	} else if circle > 0 {
		circle--
		timer.item.circle = circle
		timer.item.diff = tw.numSlots + pos - timer.pos
	} else {
		timer.item.removed = true
		newItem := &timingEntry{
			baseEntry: task,
			value:     timer.item.value,
		}
		tw.slots[pos].PushBack(newItem)
		tw.setTimerPosition(pos, newItem)
	}
}

func (tw *TimingWheel) onTick() {
	tw.tickedPos = (tw.tickedPos + 1) % tw.numSlots
	l := tw.slots[tw.tickedPos]
	tw.scanAndRunTasks(l)
}

func (tw *TimingWheel) removeTask(key any) {
	val, ok := tw.timers.Load(key)
	if !ok {
		return
	}

	timer := val
	timer.item.removed = true
	tw.timers.Delete(key)
}

func (tw *TimingWheel) run() {
	for {
		select {
		case <-tw.ticker.Chan():
			tw.onTick()
		case task := <-tw.setChannel:
			tw.setTask(&task)
		case key := <-tw.removeChannel:
			tw.removeTask(key)
		case task := <-tw.moveChannel:
			tw.moveTask(task)
		case fn := <-tw.drainChannel:
			tw.drainAll(fn)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimingWheel) runTasks(tasks []timingTask) {
	if len(tasks) == 0 {
		return
	}

	go func() {
		for i := range tasks {
			threading.RunSafe(func() {
				tw.execute(tasks[i].key, tasks[i].value)
			})
		}
	}()
}

func (tw *TimingWheel) scanAndRunTasks(l *list.List) {
	var tasks []timingTask

	for e := l.Front(); e != nil; {
		task := e.Value.(*timingEntry)
		if task.removed {
			next := e.Next()
			l.Remove(e)
			e = next
			continue
		} else if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		} else if task.diff > 0 {
			next := e.Next()
			l.Remove(e)
			// (tw.tickedPos+task.diff)%tw.numSlots
			// cannot be the same value of tw.tickedPos
			pos := (tw.tickedPos + task.diff) % tw.numSlots
			tw.slots[pos].PushBack(task)
			tw.setTimerPosition(pos, task)
			task.diff = 0
			e = next
			continue
		}

		tasks = append(tasks, timingTask{
			key:   task.key,
			value: task.value,
		})
		next := e.Next()
		l.Remove(e)
		tw.timers.Delete(task.key)
		e = next
	}

	tw.runTasks(tasks)
}

func (tw *TimingWheel) setTask(task *timingEntry) {
	if task.delay < tw.interval {
		task.delay = tw.interval
	}

	if val, ok := tw.timers.Load(task.key); ok {
		entry := val
		entry.item.value = task.value
		tw.moveTask(task.baseEntry)
	} else {
		pos, circle := tw.getPositionAndCircle(task.delay)
		task.circle = circle
		tw.slots[pos].PushBack(task)
		tw.setTimerPosition(pos, task)
	}
}

func (tw *TimingWheel) setTimerPosition(pos int, task *timingEntry) {
	if val, ok := tw.timers.Load(task.key); ok {
		timer := val
		timer.item = task
		timer.pos = pos
	} else {
		tw.timers.Store(task.key, &positionEntry{
			pos:  pos,
			item: task,
		})
	}
}
