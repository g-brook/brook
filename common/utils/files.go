package utils

import "os"

func FileExists(cf string) bool { // check if file exists
	_, err := os.Stat(cf)
	return err == nil

}
