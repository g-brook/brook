package cmd

import (
	"github.com/spf13/cobra"
)

var (
	restartCmdCli = &cobra.Command{
		Use:   "restart",
		Short: "Restart Brook client. Requires systemd installation and only supported on Linux",
		Run:   cliRestart,
	}

	startCmdCli = &cobra.Command{
		Use:   "start",
		Short: "Start daemon Brook client. Requires systemd installation and only supported on Linux",
		Run:   cliStart,
	}

	stopCmdCli = &cobra.Command{
		Use:   "stop",
		Short: "Stop Brook client. Requires systemd installation and only supported on Linux",
		Run:   cliStop,
	}

	statusCmdCli = &cobra.Command{
		Use:   "status",
		Short: "Get Brook client status. Requires systemd installation and only supported on Linux",
		Run:   cliStatus,
	}

	restartCmdSrv = &cobra.Command{
		Use:   "restart",
		Short: "Restart Brook server. Requires systemd installation and only supported on Linux",
		Run:   srvRestart,
	}

	stopCmdSrv = &cobra.Command{
		Use:   "stop",
		Short: "Stop Brook server.Requires systemd installation and only supported on Linux",
		Run:   srvStop,
	}
	startCmdSrv = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Brook server in background mode. Requires systemd installation and only supported on Linux.",
		Example: "./brook start  || ./brook start -c xxxx.json",
		Run:     srvStart,
	}
	statusCmdSrv = &cobra.Command{
		Use:   "status",
		Short: "Get Brook server status. Requires systemd installation and only supported on Linux",
		Run:   srvStatus,
	}

	version = &cobra.Command{
		Use:   "version",
		Short: "Get Brook server status. Requires systemd installation and only supported on Linux",
		Run:   versionFun,
	}
)

func InitClientCmd(rootCmd *cobra.Command) {
	startCmdCli.Flags().AddFlagSet(rootCmd.Flags())
	startCmdCli.PersistentFlags().AddFlagSet(rootCmd.PersistentFlags())
	rootCmd.AddCommand(restartCmdCli)
	rootCmd.AddCommand(stopCmdCli)
	rootCmd.AddCommand(startCmdCli)
	rootCmd.AddCommand(statusCmdCli)
	rootCmd.AddCommand(version)
}

func InitServerCmd(rootCmd *cobra.Command) {
	startCmdSrv.Flags().AddFlagSet(rootCmd.Flags())
	startCmdSrv.PersistentFlags().AddFlagSet(rootCmd.PersistentFlags())
	rootCmd.AddCommand(restartCmdSrv)
	rootCmd.AddCommand(stopCmdSrv)
	rootCmd.AddCommand(startCmdSrv)
	rootCmd.AddCommand(statusCmdSrv)
	rootCmd.AddCommand(version)
}
