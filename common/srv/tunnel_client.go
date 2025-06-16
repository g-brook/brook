package srv

import "io"

type TunnelClient struct {
	rw *io.Reader
}
