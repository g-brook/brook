/*
 * Copyright ©  https://github.com/zeromicro/go-zero/blob/master/core/timex/ticker.go
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
	"errors"
	"time"

	"github.com/brook/common/lang"
)

// errTimeout indicates a timeout.
var errTimeout = errors.New("timeout")

type (
	// Ticker interface wraps the Chan and Stop methods.
	Ticker interface {
		Chan() <-chan time.Time
		Stop()
	}

	// FakeTicker interface is used for unit testing.
	FakeTicker interface {
		Ticker
		Done()
		Tick()
		Wait(d time.Duration) error
	}

	fakeTicker struct {
		c    chan time.Time
		done chan lang.PlaceholderType
	}

	realTicker struct {
		*time.Ticker
	}
)

// NewTicker returns a Ticker.
func NewTicker(d time.Duration) Ticker {
	return &realTicker{
		Ticker: time.NewTicker(d),
	}
}

func (rt *realTicker) Chan() <-chan time.Time {
	return rt.C
}

// NewFakeTicker returns a FakeTicker.
func NewFakeTicker() FakeTicker {
	return &fakeTicker{
		c:    make(chan time.Time, 1),
		done: make(chan lang.PlaceholderType, 1),
	}
}

func (ft *fakeTicker) Chan() <-chan time.Time {
	return ft.c
}

func (ft *fakeTicker) Done() {
	ft.done <- lang.Placeholder
}

func (ft *fakeTicker) Stop() {
	close(ft.c)
}

func (ft *fakeTicker) Tick() {
	ft.c <- time.Now()
}

func (ft *fakeTicker) Wait(d time.Duration) error {
	select {
	case <-time.After(d):
		return errTimeout
	case <-ft.done:
		return nil
	}
}
