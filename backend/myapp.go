package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {

	// 1. 通过 flag 传参端口
	var serverPort string
	flag.StringVar(&serverPort, "port", "3000", "-port=6000 指定端口")
	flag.Parse()

	// 2. 返回内容，包含端口，用以辨别请求的是哪台服务器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi from Server: ", serverPort)
	})

	// 3. 启动服务
	fmt.Println("Serve at http://localhost:" + serverPort)
	http.ListenAndServe(":"+serverPort, nil)
}
