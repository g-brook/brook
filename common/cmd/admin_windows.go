package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/pid"
)

func start(service string) {
	exe, err := os.Executable()
	if err != nil {
		log.Error(err.Error())
		return
	}
	cmd := exec.Command(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("start windows service %s", service)
}

func install(service string) {
	log.Error("Not supported command:start")
}

func stop(service string) {
	currentPid := pid.CurrentPid()
	proc, err := os.FindProcess(currentPid)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = proc.Kill()
	if err != nil {
		log.Fatal("Failed to kill process:", err)
	}
	fmt.Println("Stopped process", currentPid)
	_ = pid.DeletePidFile()
}

func status(service string) {
	cpid := pid.CurrentPid()
	if cpid == 0 {
		fmt.Println("Server is not running.")
		return
	}

	proc, err := os.FindProcess(cpid)
	if err != nil {
		fmt.Println("Server PID file exists but process not found:", cpid)
		return
	}

	// 尝试发送信号0判断是否存在（Windows上FindProcess本身不保证是否存在）
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		fmt.Println("Server PID file exists but process not found:", cpid)
	} else {
		fmt.Println("Server is running with PID", cpid)
	}
}

func restart(service string) {
	stop(service)
	start(service)
}
