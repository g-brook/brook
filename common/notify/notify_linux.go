package notify

import (
	"io"
	"net"
	"os"
	"strings"
)

func sdNotify(path, payload string) error {
	socketAddr := &net.UnixAddr{
		Net:  "unixgram",
		Name: path,
	}
	conn, err := net.DialUnix(socketAddr, nil, socketAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = io.Copy(conn, strings.NewReader(payload))
	if err != nil {
		return err
	}
	return nil
}

func notifyReadiness() error {
	val, ok := os.LookupEnv("NOTIFY_SOCKET")
	if !ok || val == "" {
		return nil
	}
	if err := sdNotify(val, "READY=1"); err != nil {
		return err
	}
	return nil
}

// notifyReloading notifies systemd that caddy is reloading its config.
func notifyReloading() error {
	val, ok := os.LookupEnv("NOTIFY_SOCKET")
	if !ok || val == "" {
		return nil
	}
	if err := sdNotify(val, "RELOADING=1"); err != nil {
		return err
	}
	return nil
}

// notifyStopping notifies systemd that caddy is stopping.
func notifyStopping() error {
	val, ok := os.LookupEnv("NOTIFY_SOCKET")
	if !ok || val == "" {
		return nil
	}
	if err := sdNotify(val, "STOPPING=1"); err != nil {
		return err
	}
	return nil
}
