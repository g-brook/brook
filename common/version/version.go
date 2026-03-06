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

const BuildVersion = "0.2.1"

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

func Banner(version string, sv string) string {
	banner := `
    __                     __  
   / /__________________  / /__
  / __  / ___/ __ \/ __ \/ //_/
 / /_/ / /  / /_/ / /_/ / ,<   
/_.___/_/   \____/\____/_/|_|  
%s

Version: v%s
Website: https://www.gbrook.cc

`
	return fmt.Sprintf(banner, sv, version)
}

func ShowBanner(version string, sv string) {
	gradientLine(Banner(version, sv), 0, 200, 255, 120, 255, 120)
}

func gradientLine(text string, startR, startG, startB, endR, endG, endB int) {

	length := len(text)

	for i, c := range text {
		r := startR + (endR-startR)*i/length
		g := startG + (endG-startG)*i/length
		b := startB + (endB-startB)*i/length
		fmt.Printf("\x1b[38;2;%d;%d;%dm%c", r, g, b, c)
	}
	fmt.Print("\x1b[0m\n")
}
