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

package logger

import (
	"sync"
	"time"
)

type WebLogger struct {
	SLogger
	Protocol string    `json:"protocol"`
	Path     string    `json:"path"`
	Host     string    `json:"host"`
	Method   string    `json:"method"`
	Status   int       `json:"status"`
	ProxyId  string    `json:"proxyId"`
	HttpId   string    `json:"httpId"`
	Time     time.Time `json:"time"`
}

var (
	httpCollectorOnce sync.Once
	httpCollector     Collector[*WebLogger]
)

func HttpCollector() Collector[*WebLogger] {
	httpCollectorOnce.Do(func() {
		httpCollector = newCollector[*WebLogger]()
	})
	return httpCollector
}
