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

package clis

import (
	"time"
)

type ClientOption func(*cOptions)

type SmuxClientOption struct {
	// enable smux client.
	Enable bool

	KeepAlive bool

	Timeout time.Duration
}

// Options
// @Description: 设置的设数.
type cOptions struct {
	KeepAlive time.Duration

	Timeout time.Duration

	PingTime time.Duration

	Smux *SmuxClientOption

	handlers []ClientHandler
}

func NewSmuxClientOption() *SmuxClientOption {
	return &SmuxClientOption{
		Enable:    true,
		KeepAlive: true,
		Timeout:   time.Second * 5,
	}
}

func clientOptions(opt ...ClientOption) *cOptions {
	o := new(cOptions)
	for _, optionsFun := range opt {
		optionsFun(o)
	}
	return o
}

func WithClientSmux(opt *SmuxClientOption) ClientOption {
	return func(c *cOptions) {
		c.Smux = opt
	}
}

func WithPingTime(t time.Duration) ClientOption {
	return func(c *cOptions) {
		c.PingTime = t
	}
}

func WithTimeout(t time.Duration) ClientOption {
	return func(c *cOptions) {
		c.Timeout = t
	}
}

func WithKeepAlive(t time.Duration) ClientOption {
	return func(c *cOptions) {
		c.KeepAlive = t
	}
}

func WithClientHandler(handler ...ClientHandler) ClientOption {
	return func(c *cOptions) {
		c.handlers = append(c.handlers, handler...)
	}
}
