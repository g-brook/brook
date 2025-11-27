package api

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/brook/common/log"
	"github.com/brook/common/notify"
	"github.com/brook/common/pid"
)

func init() {
	RegisterRoute(NewRoute("/reload", "POST"), reload)
	RegisterRoute(NewRoute("/stop", "POST"), stop)
}

func reload(*Request[AuthInfo]) *Response {
	return nil
}

func stop(*Request[AuthInfo]) *Response {
	err := notify.NotifyStopping()
	if err != nil {
		log.Error("notify.NotifyStopping error: %s", err.Error())
	}
	err = exitProcess()
	if err != nil {
		log.Error("exit process failed :%v", err)
	}
	return nil
}

func exitProcess() error {
	currentPid := pid.CurrentPid()
	if currentPid == 0 {
		log.Error("Failed to get process ID")
		return errors.New("failed to get process ID")
	}
	process, err := os.FindProcess(currentPid)
	if err != nil {
		log.Error("Failed to find process", "error", err)
		return err
	}
	// 发送终止信号
	if err := sendTermSignal(process); err != nil {
		log.Error("Failed to send terminate signal", "error", err)
		return err
	}
	// 等待进程优雅退出
	if err := waitForProcessExit(currentPid, 10*time.Second); err != nil {
		log.Warn("Process did not exit gracefully, forcing kill", "error", err)
		if killErr := process.Kill(); killErr != nil {
			return err
		}
	}
	log.Info("Brook stopped successfully", "pid", currentPid)
	// 清理 PID 文件
	if err := pid.DeletePidFile(); err != nil {
		log.Warn("Failed to clean up PID file", "error", err)
	}
	return err
}

// sendTermSignal 发送终止信号到进程
func sendTermSignal(process *os.Process) error {
	if runtime.GOOS == "windows" {
		return process.Kill()
	}
	return process.Signal(syscall.SIGTERM)
}

// waitForProcessExit 等待进程退出
func waitForProcessExit(pid int, timeout time.Duration) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return nil
		}
		<-ticker.C
	}

	return fmt.Errorf("process %d did not exit within timeout", pid)
}

func processExists(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// 向进程发送 0 信号不杀死它，只做检查
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
