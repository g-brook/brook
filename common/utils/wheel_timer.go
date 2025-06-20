package utils

import (
	"github.com/RussellLuo/timingwheel"
	"time"
)

var NewWheel *timingwheel.TimingWheel

func init() {
	NewWheel = timingwheel.NewTimingWheel(time.Millisecond, 100)
	NewWheel.Start()
}
