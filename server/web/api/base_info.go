package api

import (
	"math/rand"
	"time"

	"github.com/brook/common/log"
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
	user        string = "brook"
	charset            = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNPQRSTUVWXYZ123456789!@#$%^&*()"
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
	token := RandomPassword(32)
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

func initBrookServer(*Request[any]) *Response {
	info, err := db.Get[UserInfo](userInfoKey)
	if info != nil {
		return NewResponseFail(errs.CodeSysErr, "Failed to initialize Brook server: it is already running.")
	}
	info = &UserInfo{
		Username: user,
		Password: RandomPassword(10),
	}
	err = db.Put(userInfoKey, info)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Initialize brook server fail")
	}
	//cjqQpvKuyd
	log.Info("Initialize brook server success, and userName is: %s and password is: %s", info.Username, info.Password)
	return NewResponseSuccess(info)
}

func RandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
