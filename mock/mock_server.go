package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	fmt.Println("收到了", now)
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
	content, err := http.Dir("/Users/sixh/Documents/open_project/brook/mock/").Open("mock.html")
	if err != nil {
		http.Error(w, "无法找到mock.html文件", http.StatusNotFound)
		return
	}
	defer content.Close()

	// 设置响应头为html
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 将文件内容写入响应
	_, err = io.Copy(w, content)
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
