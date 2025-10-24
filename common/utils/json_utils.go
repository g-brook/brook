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

package utils

import (
	"encoding/json"
	"os"
)

// ReaderJson reads JSON data from a file and unmarshals it into the provided output interface
// Parameters:
//
//	file: path to the JSON file to be read
//	out: pointer to the variable where the unmarshaled data will be stored
//
// Returns:
//
//	error: any error encountered during file reading or JSON unmarshaling
func ReaderJson(file string, out interface{}) error {
	// Read the file contents
	readFile, err := os.ReadFile(file)
	if err != nil {
		// Return the error if file reading fails
		return err
	}
	// Unmarshal the JSON data into the provided interface and return the result
	return json.Unmarshal(readFile, out)
}
