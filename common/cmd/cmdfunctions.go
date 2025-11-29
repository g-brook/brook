package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/brook/common/log"
	"github.com/brook/common/notify"
	"github.com/brook/common/pid"
	"github.com/spf13/cobra"
)

func cliStart(cmd *cobra.Command, args []string) {
	//if err := notify.NotifyReloading(); err != nil {
	//	log.Error("start brook failed: %v", err)
	//	return
	//}
	//defer func() {
	//	if err := notify.NotifyReadiness(); err != nil {
	//		log.Error("start brook failed: %v", err)
	//	}
	//}()
	Start("brook-cli")
}

func cliRestart(cmd *cobra.Command, args []string) {
	Restart("brook-cli")
}

func cliStop(cmd *cobra.Command, args []string) {
	Stop("brook-cli")
}
func cliStatus(cmd *cobra.Command, args []string) {
	Status("brook-cli")
}

func sevRestart(cmd *cobra.Command, args []string) {
}

func srvStop(cmd *cobra.Command, args []string) {
	if err := ServerApi("/api/stop"); err != nil {
		log.Error("stop brook failed: %v", err)
		return
	}
}

func srvStart(cmd *cobra.Command, args []string) {
	err := notify.NotifyReloading()
	if err != nil {
		log.Error("start brook failed: %v", err)
	}
	defer func() {
		if err := notify.NotifyReadiness(); err != nil {
			log.Error("start brook failed: %v", err)
		}
	}()
}

func ServerApi(path string) error {
	origin := "http://127.0.0.1:8000"
	request, err := http.NewRequest("POST", origin+path, nil)
	if err != nil {
		return fmt.Errorf("making request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", origin)
	token := pid.CurrentCliToken()
	request.Header.Set("Authorization", token)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("tcp", "127.0.0.1:8000")
			},
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("performing request: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server responded with %d", response.StatusCode)
	}
	return nil
}
