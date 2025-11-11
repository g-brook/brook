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

package configs

import (
	"fmt"
	"time"

	"github.com/brook/common/jsonx"
	"github.com/brook/common/lang"
)

var DefServerPort = 8909

// ServerConfig
// @Description: 配置文件存储.
type ServerConfig struct {
	ServerPort int                   `json:"serverPort"`
	TunnelPort int                   `json:"tunnelPort"`
	Token      string                `json:"token"`
	EnableWeb  bool                  `json:"enableWeb"`
	WebPort    int                   `json:"webPort"`
	Tunnel     []*ServerTunnelConfig `json:"tunnel"`
	Logger     LoggerConfig          `json:"logger"`
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
	Type     lang.TunnelType   `json:"type"`
	KeyFile  string            `json:"keyfile"`
	CertFile string            `json:"certFile"`
	Http     []HttpRunnelProxy `json:"http"`
}

type HttpRunnelProxy struct {
	Id     string   `json:"id"`
	Domain string   `json:"domain"`
	Paths  []string `json:"paths"`
}

type ClientTunnelConfig struct {
	TunnelType  lang.TunnelType `json:"type"`
	Destination string          `json:"destination"`
	ProxyId     string          `json:"proxyId"`
	HttpId      string          `json:"httpId,omitempty"`
	//default 1500
	UdpSize    int `json:"udpSize,omitempty"`
	RemotePort int `json:"-"`
}

// GetServerConfig
//
//	@Description: Get Config.
//	@return ServerConfig
func GetServerConfig(cfgPath string) (ServerConfig, error) {
	var cfg ServerConfig
	err := jsonx.ReaderJson(cfgPath, &cfg)
	if err != nil {
		fmt.Println(err.Error())
		return cfg, err
	}
	return cfg, nil
}

// WriterConfig
//
//	@Description: Get Config.
//	@return ServerConfig
func WriterConfig(cfgPath string, cfg *ClientConfig) error {
	err := jsonx.ReaderJson(cfgPath, cfg)
	if err != nil {
		return err
	}
	return nil
}

// ClientConfig
// @Description: Description.
type ClientConfig struct {
	ServerPort int                   `json:"serverPort"`
	ServerHost string                `json:"serverHost"`
	Token      string                `json:"token"`
	PingTime   time.Duration         `json:"pingTime"`
	Tunnels    []*ClientTunnelConfig `json:"tunnels"`
	Logger     *LoggerConfig         `json:"logger,omitempty"`
}
