package main

import (
	"log"
	"net/http"
)

// ContextKey 类型作为 r.Context().Value 的 KEY
type ContextKey string

// 配置信息
var (
	serverList = []string{
		"http://127.0.0.1:6000",
		"http://127.0.0.1:6001",
		"http://127.0.0.1:6002",
		"http://127.0.0.1:6003",
		"http://127.0.0.1:6004",
	}
	port = "8000"
)

func main() {

	// 1. 初始化连接池
	serverPool := NewServerPool(serverList)

	// 2. 转发请求
	http.HandleFunc("/", serverPool.ForwardRequest)

	// 3. 启动服务
	log.Printf("Load Balancer running at http://localhost:%v", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
