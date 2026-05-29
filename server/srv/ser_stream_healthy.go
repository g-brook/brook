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

package srv

import (
	"sync"
	"time"

	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/common/wheel"
	"github.com/xtaci/smux"
)

var schannelWheel *wheel.TimingWheel
var sessionToStream = hash.NewSyncMap[*smux.Session, *hash.SyncSet[string]]()
var idToChannel = hash.NewSyncMap[string, *transport.SChannel]()
var idToSession = hash.NewSyncMap[string, *smux.Session]()
var sessionStreamMu sync.Mutex

const (
	streamHealthyCheckInterval = 5 * time.Second
)

func init() {
	schannelWheel, _ = wheel.NewTimingWheel(100*time.Millisecond, 100, _check)
}

func addHealthyCheckStream(session *smux.Session, schannel *transport.SChannel) {
	if session == nil || schannel == nil {
		return
	}
	id := schannel.GetId()
	idToChannel.Store(id, schannel)
	idToSession.Store(id, session)
	sessionStreamMu.Lock()
	ids, ok := sessionToStream.Load(session)
	if !ok || ids == nil {
		ids = hash.NewSyncSet[string]()
		sessionToStream.Store(session, ids)
	}
	ids.Add(id)
	sessionStreamMu.Unlock()
	_ = schannelWheel.SetTimer(id, schannel, streamHealthyCheckInterval)
}

func removeHealthyCheckStream(id string) {
	if id == "" {
		return
	}
	idToChannel.Delete(id)
	session, ok := idToSession.Load(id)
	if ok && session != nil {
		sessionStreamMu.Lock()
		ids, exists := sessionToStream.Load(session)
		if exists && ids != nil {
			ids.Remove(id)
			if ids.Len() == 0 {
				sessionToStream.Delete(session)
			}
		}
		sessionStreamMu.Unlock()
	}
	idToSession.Delete(id)
}

func sessionClose(session *smux.Session) {
	if session == nil {
		return
	}
	sessionStreamMu.Lock()
	ids, ok := sessionToStream.Load(session)
	if ok {
		sessionToStream.Delete(session)
	}
	sessionStreamMu.Unlock()
	if ok && ids != nil {
		ids.ForEach(func(id string) bool {
			load, exists := idToChannel.Load(id)
			if exists && load != nil {
				_ = load.Close()
			}
			removeHealthyCheckStream(id)
			return true
		})
	}
}

func _check(_, v any) {
	if v == nil {
		return
	}
	schannel, ok := v.(*transport.SChannel)
	if !ok || schannel == nil {
		log.Warn("schannel healthy check: invalid channel type %T", v)
		return
	}
	id := schannel.GetId()
	if schannel.IsClose() {
		log.Debug("schannel healthy check: false, close")
		_ = schannel.Close()
		removeHealthyCheckStream(id)
		return
	}
	if _, exists := idToChannel.Load(id); !exists {
		log.Debug("schannel healthy check: skip stale stream %s", id)
		return
	}
	log.Debug("schannel healthy check: true")
	_ = schannelWheel.SetTimer(id, schannel, streamHealthyCheckInterval)
}
