package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// ServerPool 连接池
type ServerPool struct {
	Backends []*Server
	Current  uint64
}

const AttemptsKey ContextKey = "attempts"

func NewServerPool(servers []string) *ServerPool {

	// 1. 初始化
	serverPool := &ServerPool{
		// 初始化 Backends 的读取 index，从 0 开始
		Current: 0,
	}

	// 2. 遍历创建所有 Server 实例
	for _, serverString := range servers {
		server := NewServer(serverString, serverPool)
		serverPool.Backends = append(serverPool.Backends, server)
	}

	// 3. 返回
	return serverPool
}

// ForwardRequest 将请求迭代给连接池里的某个
func (serverPool *ServerPool) ForwardRequest(writer http.ResponseWriter, request *http.Request) {

	attempts := GetAttemptsFromContext(request)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", request.RemoteAddr, request.URL.Path)
		http.Error(writer, "Service not available", http.StatusServiceUnavailable)
		return
	}

	// 1. 获取下一个请求
	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(writer, request)
		log.Printf("Forward Request to %s, Path is %s\n", peer.URL, request.URL.Path)
		return
	}
	http.Error(writer, "No alive peer available", http.StatusServiceUnavailable)
}

// GetNextPeer 从连接池里取下一个连接，支持原子性
func (serverPool *ServerPool) GetNextPeer() *Server {
	len := len(serverPool.Backends)
	nextIdx := int(atomic.AddUint64(&serverPool.Current, uint64(1)) % uint64(len))

	// index 加 len 可以循环整个 Backends 数组
	loopCounter := nextIdx + len
	for i := nextIdx; i < loopCounter; i++ {

		// 处理 nextIdx = 4 , len = 5,  i = 6 的情况
		usedIdx := i % len
		if serverPool.Backends[usedIdx].IsAlive() {
			// 只有 nextIdx 不可用时，才需要更新 serverPool.Current 的值
			if i != nextIdx {
				atomic.StoreUint64(&serverPool.Current, uint64(usedIdx))
			}
			return serverPool.Backends[usedIdx]
		}
	}

	return nil
}

// AttemptNextServer 针对同一个请求尝试不同的后端服务，发生在服务不可用的情况
func (serverPool *ServerPool) AttemptNextServer(writer http.ResponseWriter, request *http.Request) {

	attempts := GetAttemptsFromContext(request)
	fmt.Printf("\nAttempting %s(%s) , times: %d\n\n", request.RemoteAddr, request.URL.Path, attempts)
	ctx := context.WithValue(request.Context(), AttemptsKey, attempts+1)

	serverPool.ForwardRequest(writer, request.WithContext(ctx))
}

// GetAttemptsFromContext 从 http.Request.Context 中读取 Attempts
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(AttemptsKey).(int); ok {
		return attempts
	}
	return 1
}
