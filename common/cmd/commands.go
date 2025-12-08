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
		Short: "Get Brook version.",
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
