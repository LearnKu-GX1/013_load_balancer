package main

import (
	"net/http/httputil"
	"net/url"
)

// Server 每一个 Server 对应一个后端服务 URL
type Server struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	ServerPool   *ServerPool
}

// NewServer 通过 URL 来初始化一个后端服务
func NewServer(urlStr string) *Server {

	// 1. 解析 URL
	url, _ := url.Parse(urlStr)

	// 2. 初始化后端服务
	server := &Server{
		URL: url,
	}

	// 3. 使用 httputil 包初始化后反向代理
	server.ReverseProxy = httputil.NewSingleHostReverseProxy(server.URL)

	return server
}
