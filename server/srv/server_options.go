package srv

import (
	"github.com/brook/common/utils"
	"time"
)

type ServerOption func(opts *sOptions)

// Options
// @Description: 设置的设数.
type sOptions struct {
	timeout  int64
	withSmux *SmuxServerOption
	network  utils.Network
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
	return utils.NumberDefault(t.timeout, time.Duration(30000).Milliseconds())
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
func WithNetwork(pt utils.Network) ServerOption {
	// This function takes a pointer to a sOptions struct and sets the protocol field to the value of the pt parameter
	return func(opts *sOptions) {
		opts.network = pt
	}
}
