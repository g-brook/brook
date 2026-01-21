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

package ringbuffer

import (
	"sync/atomic"

	"github.com/panjf2000/gnet/v2/pkg/buffer/ring"
)

var BufferIndex atomic.Int64

type RingBuffer struct {
	*ring.Buffer
	index int64
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		index:  BufferIndex.Add(1),
		Buffer: ring.New(size),
	}
}

func (b *RingBuffer) Index() int64 {
	return atomic.LoadInt64(&b.index)
}
