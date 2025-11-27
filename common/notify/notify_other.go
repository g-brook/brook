//go:build !linux && !windows

package notify

import (
	"errors"
)

func notifyReadiness() error { return errors.New("not support command, only support linux，windows") }
func notifyReloading() error { return errors.New("not support command, only support linux，windows") }
func notifyStopping() error  { return errors.New("not support command, only support linux，windows") }
