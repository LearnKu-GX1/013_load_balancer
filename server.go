package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Server 每一个 Server 对应一个后端服务 URL
type Server struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	ServerPool   *ServerPool

	// 是否可用的标示，false 为服务器不可用
	Alive bool

	// 读写锁，RWMutex 基于 Mutex 实现
	// RWMutex 是单写多读锁，该锁可以加多个读锁或者一个写锁
	// 读锁占用的情况下会阻止写，不会阻止读，多个 goroutine 可以同时获取读锁
	// 写锁会阻止其他 goroutine（无论读和写）进来，整个锁由该 goroutine 独占
	// 适用于读多写少的场景
	Mux sync.RWMutex
}

const RetriesKey ContextKey = "retries"

// NewServer 通过 URL 来初始化一个后端服务
func NewServer(urlStr string, serverPool *ServerPool) *Server {

	// 1. 解析 URL
	url, _ := url.Parse(urlStr)

	// 2. 初始化后端服务
	server := &Server{
		URL:        url,
		ServerPool: serverPool,
	}

	// 3. 使用 httputil 包初始化后反向代理
	server.ReverseProxy = httputil.NewSingleHostReverseProxy(server.URL)

	server.ReverseProxy.ErrorHandler = server.ProxyErrorHandler

	return server
}

// SetAlive 标记可用性
func (server *Server) SetAlive(alive bool) {
	server.Mux.Lock()
	server.Alive = alive
	server.Mux.Unlock()
}

// IsAlive 返回可用标示
func (server *Server) IsAlive() (alive bool) {
	server.Mux.RLock()
	alive = server.Alive
	server.Mux.RUnlock()
	return
}

func (server *Server) ProxyErrorHandler(writer http.ResponseWriter, request *http.Request, e error) {

	log.Printf("Proxy Error:[%s], Error %s\n", server.URL, e.Error())
	retries := GetRetryFromContext(request)
	if retries < 3 {
		log.Printf("Retry [%s] for %d times\n", server.URL, retries)
		select {
		case <-time.After(10 * time.Millisecond):
			ctx := context.WithValue(request.Context(), RetriesKey, retries+1)
			server.ReverseProxy.ServeHTTP(writer, request.WithContext(ctx))
		}
		return
	}

	ctx := context.WithValue(request.Context(), RetriesKey, 1)
	server.ServerPool.AttemptNextServer(writer, request.WithContext(ctx))
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(RetriesKey).(int); ok {
		return retry
	}
	return 0
}
