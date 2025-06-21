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

// Options defines the configuration options for a client.
type cOptions struct {
	KeepAlive time.Duration

	Timeout time.Duration

	PingTime time.Duration

	Smux *SmuxClientOption

	handlers []ClientHandler
}

// NewSmuxClientOption creates a new SmuxClientOption with default settings.
// Returns:
//   - *SmuxClientOption: A pointer to the newly created SmuxClientOption instance.
func NewSmuxClientOption() *SmuxClientOption {
	return &SmuxClientOption{
		Enable:    true,
		KeepAlive: true,
		Timeout:   time.Second * 10000,
	}
}

// clientOptions applies the provided ClientOption functions to a new cOptions instance.
// Parameters:
//   - opt: Variadic list of ClientOption functions.
//
// Returns:
//   - *cOptions: A pointer to the configured cOptions instance.
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
