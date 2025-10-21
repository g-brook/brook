package api

import (
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/brook/server/web/db"
	"github.com/brook/server/web/errs"
)

func init() {
	RegisterRoute(NewRouteNotAuth("/getBaseInfo", "POST"), getBaseInfo)
	RegisterRoute(NewRouteNotAuth("/initBrookServer", "POST"), initBrookServer)
	RegisterRoute(NewRouteNotAuth("/login", "POST"), login)
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
	token := utils.RandomString(32)
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
	bf.Version = "1.0.0"
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
