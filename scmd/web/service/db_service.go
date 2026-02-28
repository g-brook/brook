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

package service

import (
	"github.com/g-brook/brook/scmd/web/api"
	"github.com/g-brook/brook/scmd/web/db"
)

// GetToken retrieves an authentication token from the database
// It returns the token as a string or an empty string if an error occurs
func GetToken() string {
	get, err := db.Get[api.AuthInfo](api.AuthKey)
	if err != nil {
		return ""
	}
	return get.Token
}
