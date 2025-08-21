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
