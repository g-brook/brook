package srv

import (
	"github.com/RussellLuo/timingwheel"
	"time"
)

var newWheel *timingwheel.TimingWheel

func init() {
	newWheel = timingwheel.NewTimingWheel(time.Millisecond, 100)
	newWheel.Start()
}
