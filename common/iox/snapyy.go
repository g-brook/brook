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

package iox

import (
	"io"
	"sync"

	"github.com/klauspost/compress/snappy"
)

var (
	snappyRPool sync.Pool
	snappyWPool sync.Pool
)

func GetSnappyReader(r io.Reader) *snappy.Reader {
	sanppyR := snappyRPool.Get()
	if sanppyR == nil {
		return snappy.NewReader(r)
	}
	sr := sanppyR.(*snappy.Reader)
	sr.Reset(r)
	return sr
}

func PutSnappyReader(sr *snappy.Reader) {
	snappyRPool.Put(sr)
}

func GetSnappyWriter(w io.Writer) *snappy.Writer {
	sanppyW := snappyWPool.Get()
	if sanppyW == nil {
		return snappy.NewBufferedWriter(w)
	}
	sw := sanppyW.(*snappy.Writer)
	sw.Reset(w)
	return sw
}
func PutSnappyWriter(sw *snappy.Writer) {
	snappyWPool.Put(sw)
}
