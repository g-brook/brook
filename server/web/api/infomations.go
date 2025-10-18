package api

import (
	"encoding/json"

	"github.com/brook/server/web/sql"
)

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
}

type InitInfo struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type WebConfigInfo struct {
	Id         string `json:"id"`
	RefProxyId string `json:"RefProxyId"`
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
