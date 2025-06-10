package configs

import (
	"common/utils"
	"fmt"
)

var DefServerPort int32 = 8909

// ServerConfig
// @Description: 配置文件存储.
type ServerConfig struct {
	ServerPort int32 `json:"serverPort"`

	Tunnel []TunnelConfig `json:"tunnel"`

	Logger LoggerConfig `json:"logger"`
}

// LoggerConfig
// @Description:
type LoggerConfig struct {
	LoggLevel string `json:"logLevel"`

	LogPath string `json:"logPath"`
}

type TunnelConfig struct {
	//
	//  Port
	//  @Description: port.
	//
	Port int32 `json:"port"`

	Type utils.TunnelType `json:"type"`
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
// @Description: 客户端的配置信息.
type ClientConfig struct {
	ServerPort int32  `json:"serverPort"`
	ServerHost string `json:"serverHost"`
}
