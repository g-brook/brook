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

package exchange

// Heartbeat
// @Description: Ping InBound info. This is empty request,server use Cmd　discern.
type Heartbeat struct {
	Value      string `json:"value"`
	StartTime  int64  `json:"start_time"`
	ServerTime int64  `json:"server_time"`
}

// Cmd
//
//	@Description: getCmd.
//	@receiver p
//	@return Cmd
func (p Heartbeat) Cmd() Cmd {
	return Heart
}
