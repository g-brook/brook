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

func WithTimeout(timeout time.Duration) ServerOption {
	return func(opts *options) {
		opts.timeout = timeout.Milliseconds()
	}
}
