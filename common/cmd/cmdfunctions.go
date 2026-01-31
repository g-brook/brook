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
	"fmt"

	version2 "github.com/brook/common/version"
	"github.com/spf13/cobra"
)

func cliInstall(cmd *cobra.Command, args []string) {
	Install("brook-cli")
}

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
func srvInstall(cmd *cobra.Command, args []string) {
	Install("brook-sev")
}

func srvStop(cmd *cobra.Command, args []string) {
	Stop("brook-sev")
}
func srvStatus(cmd *cobra.Command, args []string) {
	Status("brook-sev")
}

func versionFun(cmd *cobra.Command, args []string) {
	fmt.Println("v" + version2.GetBuildVersion())
}
