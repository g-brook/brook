/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/g-brook/brook/common/log"
)

const (
	servicePath = "/etc/systemd/system/"
)

func install(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		printSuccess(fmt.Sprintf("Service %s %s  successfully", serviceName, "install"))
		return nil
	})
}

// start starts the specified systemd service
func start(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		job := make(chan string, 1)
		_, err := conn.StartUnitContext(context.Background(), serviceName, "replace", job)
		if err != nil {
			// Check if permission denied
			checkPermission("Start", service, err)
			return err
		}
		// Wait for the start job to complete
		<-job
		afterCheck("Start", conn, serviceName, service)
		return nil
	})
}

func stop(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		job := make(chan string, 1)
		_, err := conn.StopUnitContext(context.Background(), serviceName, "replace", job)
		if err != nil {
			// Check if permission denied
			checkPermission("Stop", service, err)
			return err
		}
		// Wait for the start job to complete
		<-job
		afterCheck("Stop", conn, serviceName, service)
		return nil
	})
}

func verifyStatus(conn *dbus.Conn, serviceName string, service string) bool {
	// Verify service has stopped
	props, err := conn.GetUnitPropertiesContext(context.Background(), serviceName)
	if err != nil {
		printError(fmt.Sprintf("Failed to verify service %s status: %v", service, err))
		return false
	}

	var activeState, subState string
	if activeVal, ok := props["ActiveState"]; ok {
		if activeStr, ok := activeVal.(string); ok {
			activeState = activeStr
		}
	}
	if subVal, ok := props["SubState"]; ok {
		if subStr, ok := subVal.(string); ok {
			subState = subStr
		}
	}

	if activeState != "" {
		if (activeState == "inactive") && (subState == "dead" || subState == "failed") {
			return false
		}
		if activeState == "failed" {
			return false
		}
	}
	return true
}

func execute(service string, fun func(conn *dbus.Conn, serviceName string) error) {
	serviceName := service + ".service"
	if !isSystemd() {
		printError("System does not support systemd, please run `systemctl daemon-reload`")
		return
	}
	conn, err := getConn(serviceName)
	if err != nil {
		printError(fmt.Sprintf("Failed to get dbus connection: %v", err))
		return
	}
	defer conn.Close()
	_ = fun(conn, serviceName)

}

func restart(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		job := make(chan string, 1)
		_, err := conn.RestartUnitContext(context.Background(), serviceName, "replace", job)
		if err != nil {
			checkPermission("Restart", service, err)
			return err
		}
		// Wait for the start job to complete
		<-job
		afterCheck("Restart", conn, serviceName, service)
		return nil
	})
}

func afterCheck(cmd string, conn *dbus.Conn, serviceName string, service string) {
	isStop := strings.ToLower(cmd) == "stop"
	if !verifyStatus(conn, serviceName, service) && !isStop {
		printError(fmt.Sprintf("Failed to %s service %s", cmd, service))
	} else {
		printSuccess(fmt.Sprintf("Service %s %s  successfully", service, cmd))
	}
}

func checkPermission(cmd, service string, err error) {
	// Check if permission denied
	if os.IsPermission(err) || strings.HasPrefix(err.Error(), "Interactive authentication required") {
		printError("Permission denied, Please run with sudo")
	} else {
		printError(fmt.Sprintf("Failed to %v service %s: %v", cmd, service, err))
	}
}
func status(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		props, err := conn.GetUnitPropertiesContext(context.Background(), serviceName)
		if err != nil {
			printError(fmt.Sprintf("Get status failed %s: %v", service, err))
			return err
		}
		var active string
		var sub string
		var pid uint32

		if activeVal, ok := props["ActiveState"]; ok {
			if activeStr, ok := activeVal.(string); ok {
				active = activeStr
			}
		}

		if subVal, ok := props["SubState"]; ok {
			if subStr, ok := subVal.(string); ok {
				sub = subStr
			}
		}

		if pidVal, ok := props["MainPID"]; ok {
			if pidUint, ok := pidVal.(uint32); ok {
				pid = pidUint
			}
		}
		fmt.Printf("%s Status: %s (%s)\n", service, active, sub)
		if pid != 0 {
			fmt.Printf("PID: %d\n", pid)
		}
		return nil
	})
}

// getConn gets a systemd dbus connection, creates service file if not exists
func getConn(serviceName string) (*dbus.Conn, error) {
	serviceFilePath := filepath.Join(servicePath, serviceName)

	// Check if service file exists
	if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
		if err := createSystemFile(serviceFilePath); err != nil {
			return nil, fmt.Errorf("failed to create service file: %w", err)
		}
	}

	conn, err := dbus.NewSystemConnectionContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}

	return conn, nil
}

// createSystemFile creates a systemd service file
func createSystemFile(serviceFilePath string) error {
	fileContent := `[Unit]
Description=Brook Tunnel Service
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
User=%s
Group=%s
ExecStart=%s

Restart=on-failure
RestartSec=5s

WorkingDirectory=%s

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=no
ReadWritePaths=%s

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target`
	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	workDir := filepath.Dir(execPath)
	runUser, runGroup := realUser()
	fileContent = fmt.Sprintf(fileContent, runUser, runGroup, execPath, workDir, workDir)

	// Create service file with secure permissions
	if err := os.WriteFile(serviceFilePath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	log.Info("Service file created successfully: %s", serviceFilePath)
	return nil
}

// isSystemd checks if the system uses systemd
func isSystemd() bool {
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}

func printError(msg string) {
	fmt.Println("\033[31m✗ ERROR:\033[0m " + msg)
}

func printSuccess(msg string) {
	fmt.Println("\033[32m✓ SUCCESS:\033[0m " + msg)
}

func realUser() (userName, groupName string) {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		u, _ := user.Lookup(sudoUser)
		g, _ := user.LookupGroupId(u.Gid)
		return u.Username, g.Name
	}

	u, _ := user.Current()
	g, _ := user.LookupGroupId(u.Gid)
	return u.Username, g.Name
}
