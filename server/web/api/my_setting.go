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

package api

import (
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/g-brook/brook/server/defin"
	"github.com/g-brook/brook/server/web/db"
	"github.com/g-brook/brook/server/web/errs"
)

func init() {
	RegisterRoute(NewRoute("/generateToken", "POST"), generateToken)
	RegisterRoute(NewRoute("/getToken", "POST"), getToken)
	RegisterRoute(NewRoute("/delToken", "POST"), delToken)
}

const (
	AuthKey = string(defin.TokenKey)
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNPQRSTUVWXYZ123456789!@#$%^&*()"
)

func generateToken(*Request[AuthInfo]) *Response {
	str := randomString(32)
	auth := AuthInfo{
		Token:      hex.EncodeToString([]byte(str)),
		Expire:     time.Now(),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:     true,
	}
	err := db.Put(AuthKey, auth)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "generate token failed")
	}
	defin.Set(defin.TokenKey, auth.Token)
	return NewResponseSuccess(auth)
}

func delToken(*Request[any]) *Response {
	err := db.Delete(AuthKey)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete token failed")
	}
	defin.Delete(defin.TokenKey)
	return NewResponseSuccess(nil)
}

func getToken(*Request[any]) *Response {
	token, _ := db.Get[AuthInfo](AuthKey)
	return NewResponseSuccess(token)
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
