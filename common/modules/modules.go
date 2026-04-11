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

package modules

import (
	"errors"
	"sync"
)

var (
	modules       = make(map[ModuleID]ModuleInfo)
	modulesToType = make(map[ModuleType][]ModuleID)
	modulesMu     sync.RWMutex
)

type ModuleID string

type ModuleType string

type Module interface {
	Module() ModuleInfo
}

const (
	TunnelPluginsModule = ModuleType("tunnelPlugins")

	ConfigsModule = ModuleType("configs")
)

type ModuleInfo struct {
	ID ModuleID

	ModuleType ModuleType

	New func() Module
}

func RegisterModule(instance Module) {
	mod := instance.Module()
	if mod.ID == "" {
		panic("module ID cannot be empty")
	}
	if mod.ModuleType == "" {
		panic("moduleType cannot be empty")
	}
	if mod.New == nil {
		panic("module New function cannot be nil")
	}
	if val := mod.New(); val == nil {
		panic("module New function must be nil")
	}
	if _, ok := modules[mod.ID]; ok {
		panic("module already registered: " + string(mod.ID))
	}
	modules[mod.ID] = mod
	modulesToType[mod.ModuleType] = append(modulesToType[mod.ModuleType], mod.ID)
}

func GetModule(name ModuleID) (ModuleInfo, error) {
	modulesMu.RLock()
	defer modulesMu.RUnlock()
	if mod, ok := modules[name]; ok {
		return mod, nil
	}
	return ModuleInfo{}, errors.New("module not found: " + string(name))
}

func GetModuleByType(name ModuleType) ([]ModuleInfo, error) {
	modulesMu.RLock()
	defer modulesMu.RUnlock()
	var info []ModuleInfo
	if ids, ok := modulesToType[name]; ok {
		for _, id := range ids {
			if t, ok := modules[id]; ok {
				info = append(info, t)
			}
		}
	}
	return info, nil
}
