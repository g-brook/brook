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

package hash

import (
	"hash/fnv"
	"math"
)

func GetHash32(data string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(data))
	return int(h.Sum32()) % math.MaxInt32
}

func MapToArray[K comparable, V any](m map[K]V) []*V {
	if m == nil {
		return nil
	}
	slice := make([]*V, len(m))
	i := 0
	for _, v := range m {
		slice[i] = &v
		i++
	}
	return slice
}
