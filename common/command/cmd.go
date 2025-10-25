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

package command

import (
	"github.com/brook/common/configs"
	"github.com/spf13/cobra"
)

// RegisterClientFlags
//
//	@Description:
//	@param cmd
//	@param server
func RegisterClientFlags(cmd *cobra.Command, config configs.ClientConfig) {
	cmd.PersistentFlags().IntVarP(&config.ServerPort, "server_port", "", config.ServerPort, "help")
}

// RegisterServerFlags
//
//	@Description:
//	@param cmd
func RegisterServerFlags(cmd *cobra.Command, config configs.ServerConfig) {

}
