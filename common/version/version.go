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

package version

import (
	"fmt"
)

const BuildVersion = "0.2.0"

const Version = 2

const DBVersion = 3

// GetBuildVersion application version.
func GetBuildVersion() string {
	return BuildVersion
}

func GetVersion() int {
	return Version
}

// GetDbVersion db version.
func GetDbVersion() int {
	return DBVersion
}

func Banner(version string) string {
	banner := `
    __                     __  
   / /__________________  / /__
  / __  / ___/ __ \/ __ \/ //_/
 / /_/ / /  / /_/ / /_/ / ,<   
/_.___/_/   \____/\____/_/|_|  
           v%s
`
	return fmt.Sprintf(banner, version)
}

func ShowBanner(version string) {
	fmt.Print(Banner(version))
}
