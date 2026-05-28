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

import (
	"fmt"
)

// VisitorRegister visitor
//
//	VisitorRegister
//	@Description: Visitor register client
type VisitorRegister struct {
	ProxyId   string `json:"proxy_id"`
	Token     string `json:"token"`
	LocalPort int    `json:"local_port"`
}

func (r *VisitorRegister) String() string {
	return `{
		"proxy_id":"` + r.ProxyId + `",
		"token":"` + r.Token + `",
		"local_port":` + fmt.Sprintf("%d", r.LocalPort) + `,	}`
}

func (r *VisitorRegister) Cmd() Cmd {
	return RegisterVisitor
}
