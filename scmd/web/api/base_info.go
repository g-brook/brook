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
	"github.com/g-brook/brook/scmd/web/db"
	"github.com/g-brook/brook/scmd/web/errs"
	"github.com/g-brook/brook/scmd/web/sql"
)

func init() {
	RegisterRoute(NewRouteNotAuth("/getBaseInfo", "POST"), getBaseInfo)
	RegisterRoute(NewRouteNotAuth("/initBrookServer", "POST"), initBrookServer)
	RegisterRoute(NewRouteNotAuth("/initDatabase", "POST"), initDataBase)
	RegisterRoute(NewRouteNotAuth("/login", "POST"), login)
	RegisterRoute(NewRoute("/upgradeDb", "POST"), upgradeDb)
}

func initDataBase(r *Request[InitInfo]) *Response {
	err := sql.InitDBStruct()
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "init database error")
	}
	return NewResponseSuccess(nil)
}

func login(req *Request[LoginInfo]) *Response {
	info, err := sql.GetUserByUserId(req.Body.Username)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Login in fail.")
	}
	if info == nil {
		return NewResponseFail(errs.CodeSysErr, "Login in fail.")
	}
	if info.UserId != req.Body.Username || info.Password != stringx.Md5String(req.Body.Password) {
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
	bf.IsRunning = initComplete()
	bf.Version = version.GetBuildVersion()
	bf.IsUpgrade, _ = sql.CheckDBVersion()
	return NewResponseSuccess(bf)
}

func initComplete() bool {
	get, err := sql.GetInfoValue("init_complete")
	return err == nil && get == "success"
}

func initBrookServer(r *Request[InitInfo]) *Response {
	complete := initComplete()
	if complete {
		return NewResponseFail(errs.CodeSysErr, "Failed to initialize Brook server: it has already been initialized.")
	}
	if r.Body.Password != r.Body.ConfirmPassword {
		return NewResponseFail(errs.CodeSysErr, "Failed to initialize Brook server: password and confirm password are not the same.")
	}
	err, _ := sql.AddUser(&sql.Users{
		Password: stringx.Md5String(r.Body.Password),
		UserId:   r.Body.Username,
		IsAdmin:  true,
		Icon:     "brook_default",
	})
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Initialize brook server fail")
	}
	log.Info("Initialize brook server success, and userName is: %s", r.Body.Username)
	err = sql.AddInfoValue("init_complete", "success")
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Initialize brook server fail")
	}
	return NewResponseSuccess(nil)
}

func upgradeDb(*Request[any]) *Response {
	err := sql.UpdateTableStruct()
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Upgrade database fail")
	}
	return NewResponseSuccess(nil)
}
