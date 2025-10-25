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

package srv

import (
	"time"

	"github.com/brook/common/lang"
	"github.com/brook/common/transport"
)

type ServerOption func(opts *sOptions)

type NewChannelFunction func(ch *GChannel) transport.Channel

// Options
// @Description: 设置的设数.
type sOptions struct {
	timeout        int64
	withSmux       *SmuxServerOption
	network        lang.Network
	newChannelFunc NewChannelFunction
}

// SmuxServerOption
// @Description:
type SmuxServerOption struct {
	enable bool
}

func DefaultServerSmux() *SmuxServerOption {
	return &SmuxServerOption{enable: true}
}

func serverOptions(opt ...ServerOption) *sOptions {
	o := new(sOptions)
	for _, optionsFun := range opt {
		optionsFun(o)
	}
	return o
}

func (t *sOptions) Timeout() int64 {
	return lang.NumberDefault(t.timeout, time.Duration(30000).Milliseconds())
}

// Smux
//
//	@Description: Smux渠道.
//	@receiver t
//	@return *SmuxServerOption
func (t *sOptions) Smux() *SmuxServerOption {
	return t.withSmux
}

// WithServerSmux
//
//	@Description:
//	@param smux
//	@return ServerOption
func WithServerSmux(smux *SmuxServerOption) ServerOption {
	return func(opts *sOptions) {
		opts.withSmux = smux
	}
}

func WithNewChannelFunc(nfunc NewChannelFunction) ServerOption {
	return func(opts *sOptions) {
		opts.newChannelFunc = nfunc
	}
}

// WithTimeout
//
//	@Description:
//	@param timeout
//	@return ServerOption
func WithTimeout(timeout time.Duration) ServerOption {
	return func(opts *sOptions) {
		opts.timeout = timeout.Milliseconds()
	}
}

// WithNetwork WithProtocol This function takes a string parameter and returns a ServerOption
func WithNetwork(pt lang.Network) ServerOption {
	// This function takes a pointer to a sOptions struct and sets the protocol field to the value of the pt parameter
	return func(opts *sOptions) {
		opts.network = pt
	}
}
