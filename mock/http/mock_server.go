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
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	file, err := os.Open("/Users/sixh/Documents/open_project/brook/mock/http/mock.html")
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

const (
	username = "admin"
	password = "123456"
)

func handlerAdmin(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")

	// 如果没有认证信息或错误，返回 401 并提示浏览器弹出认证框
	if auth == "" || !checkAuth(auth) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted Area"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// 验证通过后返回内容
	_, _ = fmt.Fprintln(w, "✅ Authenticated successfully!")
}

func checkAuth(authHeader string) bool {
	// 例如 "Basic YWRtaW46MTIzNDU2"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return false
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(data), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[0] == username && pair[1] == password
}

func main() {
	http.HandleFunc("/proxy1", mockHandler)
	http.HandleFunc("/base", handlerAdmin)
	http.HandleFunc("/proxy2", handler)
	// 启动服务器，监听 8080 端口
	fmt.Println("服务器已启动：http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		panic(err)
	}
}
