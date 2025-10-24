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

package tunnel

import (
	"fmt"
	"sync"
	"time"

	"github.com/brook/common/log"
	"github.com/brook/common/transport"
)

type GetFunction = func() error

type CheckHealthFunc func(channel transport.Channel) bool

type TunnelPool struct {
	channels        chan transport.Channel
	factory         GetFunction
	size            int
	currentSize     int
	checkHealthFunc CheckHealthFunc
	mu              sync.Mutex
}

var NewTunnelPool = func(factory GetFunction, size int) *TunnelPool {
	return &TunnelPool{
		channels: make(chan transport.Channel, size),
		size:     size,
		factory:  factory,
	}
}

func (r *TunnelPool) Get() (sch transport.Channel, err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("tunnel pool get panic", err)
		}
	}()
	var ok bool
	select {
	case sch, ok = <-r.channels:
		if ok {
			if r.checkHealthFunc != nil && !r.checkHealthFunc(sch) {
				_ = sch.Close()
				return
			}
			return
		}
	default:

	}
	err = r.factory()
	if err != nil {
		log.Error("tunnel pool get error: %v", err)
		return nil, err
	}
	select {
	case sch, ok = <-r.channels:
		if !ok {
			return nil, fmt.Errorf("tunnel pool get error: %v", err)
		}
	case <-time.After(5 * time.Second):
		log.Debug("get user tunnel timeout, 5s")
		return nil, fmt.Errorf("tunnel pool get timeout")
	}
	return
}

// Put This function takes a pointer to a transport.SChannel and puts it into a channel
func (r *TunnelPool) Put(sch transport.Channel) error {
	// This deferred function will be called when the function returns
	defer func() {
		// If there is an error, it will be recovered and logged
		if err := recover(); err != nil {
			log.Error("tunnel pool put panic", err)
		}
	}()
	// This select statement will put the SChannel into the channel
	select {
	case r.channels <- sch:
		log.Debug("tunnel pool connection registered")
		return nil
	default:
		r.mu.Lock()
		_ = sch.Close()
		r.currentSize--
		r.mu.Unlock()
		log.Debug("tunnel pool put error")
		return fmt.Errorf("tunnel pool put error")
	}
}
