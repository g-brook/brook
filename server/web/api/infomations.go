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
	"encoding/json"
	"time"

	"github.com/brook/server/web/sql"
)

type AuthInfo struct {
	Token      string `json:"token"`
	Status     bool   `json:"status"`
	CreateTime string `json:"createTime"`
	//过期时间
	Expire time.Time `json:"expire"`
}

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BaseInfo struct {
	IsRunning bool   `json:"isRunning"`
	Version   string `json:"version"`
}

type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type QueryServerInfo struct {
	Name    string `json:"name"`
	Port    string `json:"port"`
	ProxyId string `json:"proxyId"`
}

type ServerInfo struct {
	Name        string `json:"name"`
	Port        string `json:"port"`
	TunnelType  string `json:"tunnelType"`
	TAG         string `json:"tag"`
	Connections int    `json:"connections"`
	Users       int    `json:"users"`
	ProxyId     string `json:"proxyId"`
}

type InitInfo struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type ServerClientInfo struct {
	Host     string `json:"host"`
	LastTime string `json:"lastTime"`
}

type WebConfigInfo struct {
	Id         string `json:"id"`
	RefProxyId int    `json:"RefProxyId"`
	CertFile   string `json:"certFile"`
	KeyFile    string `json:"keyFile"`
	Proxy      []struct {
		Id     string   `json:"id"`
		Domain string   `json:"domain"`
		Paths  []string `json:"paths"`
	} `json:"proxy"`
}

func (r WebConfigInfo) toDb() sql.WebProxyConfig {
	j, _ := json.Marshal(r.Proxy)
	return sql.WebProxyConfig{
		Id:         r.Id,
		RefProxyId: r.RefProxyId,
		CertFile:   r.CertFile,
		KeyFile:    r.KeyFile,
		Proxy:      string(j),
	}
}
