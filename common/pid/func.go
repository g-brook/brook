package pid

import (
	"fmt"
	"os"
	"strconv"

	"github.com/brook/common/log"
)

var (
	File      = "pid"
	TokenFile = "cli_token"
)

func CurrentPid() int {
	file, err := os.ReadFile(File)
	if err != nil {
		log.Error("Failed to read PID file: %v", err)
		return 0
	}
	currentPid, err := strconv.Atoi(string(file))
	if err != nil {
		return 0
	}
	return currentPid
}

func CreatePidFile() {
	// 写入PID文件
	pidFile := File
	currentPid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", currentPid)), 0777); err != nil {
		log.Error("Failed to write PID file: %v\n", err)
	} else {
		log.Info("PID file created: %s, PID: %d\n", pidFile, currentPid)
	}
}

func DeletePidFile() error {
	if err := os.Remove(File); err != nil {
		return err
	} else {
		log.Info("PID file removed: %s\n", File)
		return nil
	}
}

func CurrentCliToken() string {
	file, err := os.ReadFile(TokenFile)
	if err != nil {
		log.Error("Failed to read token file: %v", err)
		return ""
	}
	return string(file)
}

func CreateCliTokenFile(token string) {
	if err := os.WriteFile(TokenFile, []byte(token), 0777); err != nil {
		log.Error("Failed to write token file: %v\n", err)
	} else {
		log.Info("Token file created token: %d\n", token)
	}
}
