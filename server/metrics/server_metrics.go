package metrics

import (
	"time"
)

type TunnelMetrics interface {
	Id() string
	Name() string
	Port() int
	Type() string
	Connections() int
	Users() int
	Runtime() time.Time
}
