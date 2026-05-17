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

package remote

import (
	"time"

	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/wheel"
	"github.com/g-brook/brook/server/srv"
)

var gchannelWheel *wheel.TimingWheel

const (
	gchannelHealthyCheckInterval = 30 * time.Second
)

func init() {
	gchannelWheel, _ = wheel.NewTimingWheel(100*time.Millisecond, 100, _gcheck)
}

func addHealthyCheckChannel(gchannel *srv.GChannel) {
	if gchannel == nil {
		return
	}
	_ = gchannelWheel.SetTimer(gchannel.GetId(), gchannel, gchannelHealthyCheckInterval)
}

func _gcheck(_, v any) {
	if v != nil {
		gchannel := v.(*srv.GChannel)
		if gchannel.IsClose() || !gchannel.IsHealthy() {
			log.Debug("gchannel healthy check: false, addr:%s,close:%s:healthy:%s", gchannel.RemoteAddr(), gchannel.IsClose(), gchannel.IsHealthy())
			_ = gchannel.Close()
		} else {
			log.Debug("gchannel healthy check: true")
			_ = gchannelWheel.SetTimer(gchannel.GetId(), gchannel, gchannelHealthyCheckInterval)
		}
	}
}
