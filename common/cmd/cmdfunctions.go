package cmd

import (
	"fmt"

	version2 "github.com/brook/common/version"
	"github.com/spf13/cobra"
)

func cliStart(cmd *cobra.Command, args []string) {
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

func srvStart(cmd *cobra.Command, args []string) {
	Start("brook-sev")
}

func srvRestart(cmd *cobra.Command, args []string) {
	Restart("brook-sev")
}

func srvStop(cmd *cobra.Command, args []string) {
	Stop("brook-sev")
}
func srvStatus(cmd *cobra.Command, args []string) {
	Status("brook-sev")
}

func versionFun(cmd *cobra.Command, args []string) {
	fmt.Println(version2.GetBuildVersion())
}
