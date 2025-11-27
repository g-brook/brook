package cmd

import (
	"github.com/spf13/cobra"
)

var (
	restartCmdCli = &cobra.Command{
		Use:   "restart",
		Short: "Restart brook client",
		Run:   cliRestart,
	}

	startCmdCli = &cobra.Command{
		Use:   "start",
		Short: "Start daemon brook client",
		Run:   cliStart,
	}

	stopCmdCli = &cobra.Command{
		Use:   "stop",
		Short: "Stop brook client",
		Run:   srvStop,
	}

	restartCmdSrv = &cobra.Command{
		Use:   "restart",
		Short: "Restart brook server",
		Run:   sevRestart,
	}

	stopCmdSrv = &cobra.Command{
		Use:   "stop",
		Short: "Stop brook server",
		Run:   srvStop,
	}
	startCmdSrv = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Brook server in background mode. Supported on Linux and Windows only.",
		Example: "./brook start  || ./brook start -c xxxx.json",
		Run:     srvStart,
	}
)

func InitClientCmd(rootCmd *cobra.Command) {
	startCmdCli.Flags().AddFlagSet(rootCmd.Flags())
	startCmdCli.PersistentFlags().AddFlagSet(rootCmd.PersistentFlags())
	rootCmd.AddCommand(restartCmdCli)
	rootCmd.AddCommand(stopCmdCli)
	rootCmd.AddCommand(startCmdCli)
}

func InitServerCmd(rootCmd *cobra.Command) {
	restartCmdSrv.Flags().AddFlagSet(rootCmd.Flags())
	restartCmdSrv.PersistentFlags().AddFlagSet(rootCmd.PersistentFlags())
	rootCmd.AddCommand(restartCmdSrv)
	rootCmd.AddCommand(stopCmdSrv)
	rootCmd.AddCommand(startCmdSrv)
}
