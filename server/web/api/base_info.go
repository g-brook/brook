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
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/stringx"
	"github.com/g-brook/brook/common/version"
	"github.com/g-brook/brook/server/web/db"
	"github.com/g-brook/brook/server/web/errs"
	"github.com/g-brook/brook/server/web/sql"
)

func init() {
	RegisterRoute(NewRouteNotAuth("/getBaseInfo", "POST"), getBaseInfo)
	RegisterRoute(NewRouteNotAuth("/initBrookServer", "POST"), initBrookServer)
	RegisterRoute(NewRouteNotAuth("/login", "POST"), login)
	RegisterRoute(NewRoute("/upgradeDb", "POST"), upgradeDb)
}

const (
	userInfoKey string = "brook_user_info"
)

func login(req *Request[LoginInfo]) *Response {
	info, err := db.Get[UserInfo](userInfoKey)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Login in fail.")
	}
	if info == nil {
		return NewResponseFail(errs.CodeSysErr, "Login in fail.")
	}
	if info.Username != req.Body.Username || info.Password != req.Body.Password {
		return NewResponseFail(errs.CodeSysErr, "Login in fail. Username or password is wrong.")
	}
	token := stringx.RandomString(32)
	err = db.PutWithTtl(token, info, TokenTtl)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Login in fail.")
	}
	return NewResponseSuccess(token)
}

func getBaseInfo(*Request[any]) *Response {
	bf := new(BaseInfo)
	get, err := db.Get[UserInfo](userInfoKey)
	bf.IsRunning = err == nil && get != nil
	bf.Version = version.GetBuildVersion()
	bf.IsUpgrade, _ = sql.CheckDBVersion()
	return NewResponseSuccess(bf)
}

func initBrookServer(r *Request[InitInfo]) *Response {
	info, err := db.Get[UserInfo](userInfoKey)
	if info != nil {
		return NewResponseFail(errs.CodeSysErr, "Failed to initialize Brook server: it has already been initialized.")
	}
	if r.Body.Password != r.Body.ConfirmPassword {
		return NewResponseFail(errs.CodeSysErr, "Failed to initialize Brook server: password and confirm password are not the same.")
	}
	info = &UserInfo{
		Username: r.Body.Username,
		Password: r.Body.Password,
	}
	err = db.Put(userInfoKey, info)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Initialize brook server fail")
	}
	log.Info("Initialize brook server success, and userName is: %s and password is: %s", info.Username, info.Password)
	return NewResponseSuccess(info)
}

func upgradeDb(*Request[any]) *Response {
	err := sql.UpdateTableStruct()
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Upgrade database fail")
	}
	return NewResponseSuccess(nil)
}
