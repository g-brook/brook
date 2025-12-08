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
	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return err
	}
	defer func(conn *net.UnixConn) {
		_ = conn.Close()
	}(conn)
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
