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
	sql2 "database/sql"
	"encoding/json"
	"time"

	"github.com/brook/common/transform"
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
	IsUpgrade bool   `json:"isUpgrade"`
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
	Name        string    `json:"name"`
	Port        string    `json:"port"`
	TunnelType  string    `json:"tunnelType"`
	TAG         string    `json:"tag"`
	Connections int       `json:"connections"`
	Users       int       `json:"users"`
	ProxyId     string    `json:"proxyId"`
	Runtime     time.Time `json:"runtime"`
}

type InitInfo struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type ServerClientInfo struct {
	Host     string `json:"host"`
	LastTime string `json:"lastTime"`
	AgentId  string `json:"agentId"`
}

type WebConfigInfo struct {
	Id         string `json:"id"`
	RefProxyId int    `json:"RefProxyId"`
	CertId     *int   `json:"certId"`
	Proxy      []struct {
		Id     string   `json:"id"`
		Domain string   `json:"domain"`
		Paths  []string `json:"paths"`
	} `json:"proxy"`
}

type ProxyConfig struct {
	Idx         int    `json:"id"`
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	RemotePort  int    `json:"remotePort"`
	ProxyID     string `json:"proxyId"`
	Protocol    string `json:"protocol"`
	State       int    `json:"state"`
	Destination string `json:"destination"`
	IsRunning   bool   `json:"isRunning"`
	Runtime     string `json:"runtime"`
	IsExistWeb  bool   `json:"isExistWeb"`
	Clients     int    `json:"clients"`
}

type Certificate struct {
	ID         int     `json:"id" maps:"id"`
	Name       string  `json:"name" maps:"name"`
	Content    string  `json:"content" maps:"content"`
	PrivateKey string  `json:"privateKey" maps:"private_key"`
	Desc       string  `json:"desc" maps:"desc"`
	ExpireTime *string `json:"expireTime" maps:"-"`
}

func (r *ProxyConfig) IsHttpOrHttps() bool {
	return r.Protocol == "HTTP" || r.Protocol == "HTTPS"
}

func (r *Certificate) toDb() *sql.Certificate {
	var target sql.Certificate
	converter := transform.NewConverter()
	err := converter.Convert(r, &target)
	if err != nil {
		return nil
	}
	return &target
}

func (r *ProxyConfig) toDb() *sql.ProxyConfig {
	return &sql.ProxyConfig{
		Idx:         r.Idx,
		Name:        r.Name,
		Tag:         r.Tag,
		RemotePort:  r.RemotePort,
		ProxyID:     r.ProxyID,
		Protocol:    r.Protocol,
		State:       r.State,
		Destination: sql2.NullString{String: r.Destination},
	}
}
func newProxyConfig(config *sql.ProxyConfig) *ProxyConfig {
	return &ProxyConfig{
		Idx:         config.Idx,
		Name:        config.Name,
		Tag:         config.Tag,
		RemotePort:  config.RemotePort,
		ProxyID:     config.ProxyID,
		Protocol:    config.Protocol,
		State:       config.State,
		Destination: config.Destination.String,
	}
}

func (r WebConfigInfo) toDb() *sql.WebProxyConfig {
	j, _ := json.Marshal(r.Proxy)
	var nullInt32 sql2.NullInt32
	if r.CertId == nil {
		nullInt32 = sql2.NullInt32{
			Valid: false,
		}
	} else {
		nullInt32 = sql2.NullInt32{
			Valid: true,
			Int32: int32(*r.CertId),
		}
	}
	return &sql.WebProxyConfig{
		Id:         r.Id,
		RefProxyId: r.RefProxyId,
		CertId:     nullInt32,
		Proxy:      string(j),
	}
}
