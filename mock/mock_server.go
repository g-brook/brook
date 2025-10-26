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

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// 处理错误
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	_, err = fmt.Fprintf(w, "OK:%v", now)
	if err != nil {
		return
	}

}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	// 读取mock.html文件内容
	file, err := os.Open("mock.html")
	if err != nil {
		http.Error(w, "无法找到mock.html文件", http.StatusNotFound)
		return
	}
	defer file.Close()

	// 设置响应头为html
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 将文件内容写入响应
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "读取文件失败", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/proxy1", mockHandler)
	http.HandleFunc("/proxy2", handler)
	// 启动服务器，监听 8080 端口
	fmt.Println("服务器已启动：http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
