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

package pid

import (
	"fmt"
	"os"
	"strconv"

	"github.com/g-brook/brook/common/log"
)

var (
	File = "pid"
)

func CurrentPid() int {
	file, err := os.ReadFile(File)
	if err != nil {
		log.Error("Failed to read PID file: %v", err)
		return 0
	}
	currentPid, err := strconv.Atoi(string(file))
	if err != nil {
		return 0
	}
	return currentPid
}

func CreatePidFile() {
	// 写入PID文件
	pidFile := File
	currentPid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", currentPid)), 0777); err != nil {
		log.Error("Failed to write PID file: %v\n", err)
	} else {
		log.Info("PID file created: %s, PID: %d\n", pidFile, currentPid)
	}
}

func DeletePidFile() error {
	if err := os.Remove(File); err != nil {
		return err
	} else {
		log.Info("PID file removed: %s\n", File)
		return nil
	}
}
