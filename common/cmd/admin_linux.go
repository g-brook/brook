package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/brook/common/log"
	"github.com/coreos/go-systemd/v22/dbus"
)

const (
	servicePath = "/etc/systemd/system/"
)

// start starts the specified systemd service
func start(service string) {
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

	job := make(chan string, 1)
	_, err = conn.StartUnitContext(context.Background(), serviceName, "replace", job)
	if err != nil {
		log.Error("Failed to start service %s: %v", service, err)
		return
	}
	// Wait for the start job to complete
	<-job
	log.Info("Service %s started successfully", service)
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
