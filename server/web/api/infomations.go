package api

import (
	"github.com/brook/common/utils"
)

type ConfigInfo struct {
	Id      string           `json:"id"`
	Name    string           `json:"name"`
	Tag     string           `json:"tag"`
	Port    int              `json:"port"`
	ProxyId string           `json:"proxy_id"`
	Type    utils.TunnelType `json:"type"`
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
	Name string `json:"name"`
	Port string `json:"port"`
}

type ServerInfo struct {
	Name        string `json:"name"`
	Port        string `json:"port"`
	TunnelType  string `json:"tunnelType"`
	TAG         string `json:"tag"`
	Connections int    `json:"connections"`
	Users       int    `json:"users"`
}
