package defin

import (
	"fmt"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"io"
)

type DuplexBuffer struct {
	BufferPool *utils.ByteBufPool
	BufferSize int
}

func NewDuplexBuffer() *DuplexBuffer {
	return &DuplexBuffer{
		BufferSize: 32 * 1024,
		BufferPool: utils.GetBuffPool32k(),
	}
}

func (d *DuplexBuffer) Copy(a, b transport.Channel) {
	go d.pipe("b->a", b, a)
	go d.pipe("a->b", a, b)
}

func (d *DuplexBuffer) pipe(dir string, src transport.Channel, dest transport.Channel) {
	fmt.Println(dir)
	for {
		buf := d.BufferPool.Get()
		n, err := io.ReadFull(src, buf)
		if err == io.EOF {
			return
		} else {
			if n > 0 {
				data := buf[:n]
				if _, err := dest.Write(data); err != nil {
					log.Info("pipe error: %v", err)
					return
				}
			}
		}
		d.BufferPool.Put(buf)
	}
}
