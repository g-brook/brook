package remote

import "time"

type ClientOption func(*cOptions)

// Options
// @Description: 设置的设数.
type cOptions struct {
	KeepAlive time.Duration

	Timeout time.Duration
}

func clientOptions(opt ...ClientOption) *cOptions {
	o := new(cOptions)
	for _, optionsFun := range opt {
		optionsFun(o)
	}
	return o
}
