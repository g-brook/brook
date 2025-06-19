package srv

import "time"

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
		Timeout:   time.Second * 10000,
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

func WithClientHandler(handler ...ClientHandler) ClientOption {
	return func(c *cOptions) {
		c.handlers = append(c.handlers, handler...)
	}
}
