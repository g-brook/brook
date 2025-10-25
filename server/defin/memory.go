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

package defin

import (
	"sync"

	"github.com/brook/common/lang"
)

var Store Memory

func init() {
	Store = Memory{
		values: make(map[lang.KeyType]interface{}),
	}
}

type Memory struct {
	values map[lang.KeyType]any
	rn     sync.RWMutex
}

func Get[T any](key lang.KeyType) T {
	a := Store.values[key]
	return a.(T)
}

func Set(key lang.KeyType, value any) {
	Store.rn.Lock()
	defer Store.rn.Unlock()
	Store.values[key] = value
}

func Delete(key lang.KeyType) {
	Store.rn.Lock()
	defer Store.rn.Unlock()
	delete(Store.values, key)
}

func GetToken() string {
	return Get[string](TokenKey)
}
