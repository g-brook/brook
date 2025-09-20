package configs

import (
	"fmt"
	"time"

	"github.com/brook/common/utils"
)

var DefServerPort = 8909

// ServerConfig
// @Description: 配置文件存储.
type ServerConfig struct {
	ServerPort int                  `json:"serverPort"`
	TunnelPort int                  `json:"tunnelPort"`
	Tunnel     []ServerTunnelConfig `json:"tunnel"`
	Logger     LoggerConfig         `json:"logger"`
}

// LoggerConfig
// @Description:
type LoggerConfig struct {
	LoggLevel string `json:"logLevel"`
	LogPath   string `json:"logPath"`
	Outs      string `json:"outs"`
}

type ServerTunnelConfig struct {
	Id       string            `json:"id"`
	Port     int               `json:"port"`
	Type     utils.TunnelType  `json:"type"`
	KeyFile  string            `json:"keyfile"`
	CertFile string            `json:"certFile"`
	Proxy    []HttpRunnelProxy `json:"proxy"`
}

type HttpRunnelProxy struct {
	Id     string   `json:"id"`
	Domain string   `json:"domain"`
	Paths  []string `json:"paths"`
}

type ClientTunnelConfig struct {
	TunnelType   utils.TunnelType `json:"type"`
	LocalAddress string           `json:"localAddress"`
	RemotePort   int              `json:"remotePort"`
	ProxyId      string           `json:"ProxyId"`
	//default 1500
	UdpSize int `json:"udpSize"`
	Network utils.Network
}

// GetServerConfig
//
//	@Description: Get Config.
//	@return ServerConfig
func GetServerConfig(cfgPath string) (ServerConfig, error) {
	var cfg ServerConfig
	err := utils.ReaderJson(cfgPath, &cfg)
	if err != nil {
		fmt.Println(err.Error())
		return cfg, err
	}
	return cfg, nil
}

// GetClientConfig
//
//	@Description: Get Config.
//	@return ServerConfig
func GetClientConfig(cfgPath string) (ClientConfig, error) {
	var cfg ClientConfig
	err := utils.ReaderJson(cfgPath, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ClientConfig
// @Description: Description.
type ClientConfig struct {
	ServerPort int                   `json:"serverPort"`
	ServerHost string                `json:"serverHost"`
	PingTime   time.Duration         `json:"pingTime"`
	Tunnels    []*ClientTunnelConfig `json:"tunnels"`
	Logger     LoggerConfig          `json:"logger"`
}
