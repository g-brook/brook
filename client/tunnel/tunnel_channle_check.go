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
	"time"

	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/wheel"
)

var schannelWheel *wheel.TimingWheel

const (
	streamHealthyCheckInterval = 300 * time.Second
)

func init() {
	schannelWheel, _ = wheel.NewTimingWheel(100*time.Millisecond, 100, allCheck)
}
func addHealthyCheckStream(schannel *transport.SChannel) {
	if schannel == nil {
		return
	}
	_ = schannelWheel.SetTimer(schannel.GetId(), schannel, streamHealthyCheckInterval)
}

func allCheck(_, v any) {
	if v != nil {
		schannel := v.(*transport.SChannel)
		if schannel.IsClose() {
			log.Debug("schannel healthy check: false 1, close")
			_ = schannel.Close()
			return
		}
		if !schannel.IsHealthy() {
			log.Debug("schannel healthy check: false 2, close")
			_ = schannel.Close()
			return
		}
		log.Debug("schannel healthy check: true")
		_ = schannelWheel.SetTimer(schannel.GetId(), schannel, streamHealthyCheckInterval)
	}
}
