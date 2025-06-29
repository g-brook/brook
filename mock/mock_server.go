package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("收到了", time.Now())
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// 处理错误
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	_, err = fmt.Fprintf(w, "Hello, World! You accessed: %s", time.Now())
	if err != nil {
		return
	}

}

func main() {
	http.HandleFunc("/", handler)
	// 启动服务器，监听 8080 端口
	fmt.Println("服务器已启动：http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
