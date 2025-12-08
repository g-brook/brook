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

package threading

import (
	"errors"
	"sync"

	"github.com/brook/common/lang"
	"github.com/brook/common/resue"
)

// ErrTaskRunnerBusy is the error that indicates the runner is busy.
var ErrTaskRunnerBusy = errors.New("task runner is busy")

// A TaskRunner is used to control the concurrency of goroutines.
type TaskRunner struct {
	limitChan chan lang.PlaceholderType
	waitGroup sync.WaitGroup
}

// NewTaskRunner returns a TaskRunner.
func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{
		limitChan: make(chan lang.PlaceholderType, concurrency),
	}
}

// Schedule schedules a task to run under concurrency control.
func (rp *TaskRunner) Schedule(task func()) {
	// Why we add waitGroup first, in case of race condition on starting a task and wait returns.
	// For example, limitChan is full, and the task is scheduled to run, but the waitGroup is not added,
	// then the wait returns, and the task is then scheduled to run, but caller thinks all tasks are done.
	// the same reason for ScheduleImmediately.
	rp.waitGroup.Add(1)
	rp.limitChan <- lang.Placeholder
	GoSafe(func() {
		defer resue.Recover(func() {
			<-rp.limitChan
			rp.waitGroup.Done()
		})

		task()
	})
}

// ScheduleImmediately schedules a task to run immediately under concurrency control.
// It returns ErrTaskRunnerBusy if the runner is busy.
func (rp *TaskRunner) ScheduleImmediately(task func()) error {
	// Why we add waitGroup first, check the comment in Schedule.
	rp.waitGroup.Add(1)
	select {
	case rp.limitChan <- lang.Placeholder:
	default:
		rp.waitGroup.Done()
		return ErrTaskRunnerBusy
	}
	GoSafe(func() {
		defer resue.Recover(func() {
			<-rp.limitChan
			rp.waitGroup.Done()
		})
		task()
	})
	return nil
}

// Wait waits all running tasks to be done.
func (rp *TaskRunner) Wait() {
	rp.waitGroup.Wait()
}
