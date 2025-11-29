package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/brook/common/log"
	"github.com/coreos/go-systemd/v22/dbus"
)

const (
	servicePath = "/etc/systemd/system/"
)

// start starts the specified systemd service
func start(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		job := make(chan string, 1)
		_, err := conn.StopUnitContext(context.Background(), serviceName, "replace", job)
		if err != nil {
			// Check if permission denied
			if os.IsPermission(err) || strings.HasPrefix(err.Error(), "Interactive authentication required") {
				fmt.Println("Permission denied, Please run with sudo")
			} else {
				log.Error("Failed to Started service %s: %v", service, err)
			}
			return err
		}
		// Wait for the start job to complete
		<-job
		if !verifyStatus(conn, serviceName, service) {
			log.Error("Failed to start service %s: %v", service, err)
		} else {
			log.Info("Service %s Started successfully", service)
		}
		return nil
	})
}

func stop(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		job := make(chan string, 1)
		_, err := conn.StopUnitContext(context.Background(), serviceName, "replace", job)
		if err != nil {
			// Check if permission denied
			if os.IsPermission(err) || strings.HasPrefix(err.Error(), "Interactive authentication required") {
				fmt.Println("Permission denied, Please run with sudo")
			} else {
				log.Error("Failed to stop service %s: %v", service, err)
			}
			return err
		}
		// Wait for the start job to complete
		<-job
		if !verifyStatus(conn, serviceName, service) {
			log.Error("Failed to stop service %s: %v", service, err)
		} else {
			log.Info("Service %s Stopped successfully", service)
		}
		return nil
	})
}

func verifyStatus(conn *dbus.Conn, serviceName string, service string) bool {
	// Verify service has stopped
	props, err := conn.GetUnitPropertiesContext(context.Background(), serviceName)
	if err != nil {
		log.Error("Failed to verify service %s status: %v", service, err)
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
		log.Error("System does not support systemd, please run `systemctl daemon-reload`")
		return
	}
	conn, err := getConn(serviceName)
	if err != nil {
		log.Error("Failed to get dbus connection: %v", err)
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
			// Check if permission denied
			if os.IsPermission(err) || strings.HasPrefix(err.Error(), "Interactive authentication required") {
				fmt.Println("Permission denied, Please run with sudo")
			} else {
				log.Error("Failed to Restart service %s: %v", service, err)
			}
			return err
		}
		// Wait for the start job to complete
		<-job
		if !verifyStatus(conn, serviceName, service) {
			log.Error("Failed to restart service %s: %v", service, err)
		} else {
			log.Info("Service %s Restart successfully", service)
		}
		return nil
	})
}
func status(service string) {
	execute(service, func(conn *dbus.Conn, serviceName string) error {
		props, err := conn.GetUnitPropertiesContext(context.Background(), serviceName)
		if err != nil {
			// Check if permission denied
			if os.IsPermission(err) || strings.HasPrefix(err.Error(), "Interactive authentication required") {
				fmt.Println("Permission denied, Please run with sudo")
			} else {
				log.Error("Get status failed %s: %v", service, err)
			}
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
		fmt.Printf("Status: %s (%s)\n", active, sub)
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
ExecStart=%s
Restart=on-failure
RestartSec=5s
WorkingDirectory=%s

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=%s

[Install]
WantedBy=multi-user.target`

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	workDir := filepath.Dir(execPath)
	fileContent = fmt.Sprintf(fileContent, execPath, workDir, workDir)

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
