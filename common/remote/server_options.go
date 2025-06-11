package remote

import (
	"common/utils"
	"time"
)

type ServerOption func(opts *options)

// Options
// @Description: 设置的设数.
type options struct {
	timeout int64

	withSmux *SmuxOption
}

// SmuxOption
// @Description:
type SmuxOption struct {
	enable bool
}

func DefaulServerSmux() *SmuxOption {
	return &SmuxOption{enable: true}
}

func NewSmuxOption() {

}

func loadOptions(opt ...ServerOption) *options {
	o := new(options)
	for _, optionsFun := range opt {
		optionsFun(o)
	}
	return o
}

func (t *options) Timeout() int64 {
	return utils.NumberDefault(t.timeout, time.Duration(30000).Milliseconds())
}

// Smux
//
//	@Description: Smux渠道.
//	@receiver t
//	@return *SmuxOption
func (t *options) Smux() *SmuxOption {
	return t.withSmux
}

// WithSmun
//
//	@Description:
//	@param smux
//	@return ServerOption
func WithSmun(smux *SmuxOption) ServerOption {
	return func(opts *options) {
		opts.withSmux = smux
	}
}

// WithTimeout
//
//	@Description:
//	@param timeout
//	@return ServerOption
func WithTimeout(timeout time.Duration) ServerOption {
	return func(opts *options) {
		opts.timeout = timeout.Milliseconds()
	}
}
