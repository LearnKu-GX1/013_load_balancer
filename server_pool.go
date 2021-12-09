package main

import (
	"net/http"
)

// ServerPool 连接池
type ServerPool struct {
	Backends []*Server
	Current  int
}

func NewServerPool(servers []string) *ServerPool {

	// 1. 初始化
	serverPool := &ServerPool{
		// 初始化从 0 开始
		Current: 0,
	}

	// 2. 遍历创建所有 Server 实例
	for _, serverString := range servers {
		server := NewServer(serverString)
		serverPool.Backends = append(serverPool.Backends, server)
	}

	// 3. 返回
	return serverPool
}

// ForwardRequest 将请求迭代给连接池里的某个
func (serverPool *ServerPool) ForwardRequest(writer http.ResponseWriter, request *http.Request) {

	// 1. 获取下一个请求
	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(writer, request)
		return
	}
	http.Error(writer, "Service not available", http.StatusServiceUnavailable)
}

// GetNextPeer 从连接池里取下一个连接
func (serverPool *ServerPool) GetNextPeer() *Server {
	serverPool.Current = (serverPool.Current + 1) % len(serverPool.Backends)
	return serverPool.Backends[serverPool.Current]
}
