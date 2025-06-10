package command

import (
	"common/configs"
	"github.com/spf13/cobra"
)

// RegisterClientFlags
//
//	@Description:
//	@param cmd
//	@param server
func RegisterClientFlags(cmd *cobra.Command, config configs.ClientConfig) {
	cmd.PersistentFlags().Int32VarP(&config.ServerPort, "server_port", "", configs.DefServerPort, "help")
}

// RegisterServerFlags
//
//	@Description:
//	@param cmd
func RegisterServerFlags(cmd *cobra.Command, config configs.ServerConfig) {

}
