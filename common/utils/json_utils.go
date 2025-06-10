package utils

import (
	"encoding/json"
	"os"
)

// ReaderJson
//
//	@Description: 获取一个yaml对像.
//	@param file 文件名.
//	@param out 输出的对象.
//	@return error error.
func ReaderJson(file string, out interface{}) error {

	readFile, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(readFile, out)
}
