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
		Alive:      true,
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
	retries := server.getRetryFromContext(request)
	if retries < 3 {
		log.Printf("Retry [%s] for %d times\n", server.URL, retries)

		// 休息 10 毫秒（千分之一秒），给后端一点点恢复的时间
		time.Sleep(10 * time.Millisecond)
		ctx := context.WithValue(request.Context(), RetriesKey, retries+1)
		server.ReverseProxy.ServeHTTP(writer, request.WithContext(ctx))

		return
	}

	server.SetAlive(false)

	ctx := context.WithValue(request.Context(), RetriesKey, 1)
	server.ServerPool.AttemptNextServer(writer, request.WithContext(ctx))
}

// ReachableCheck 检测后端服务是否可用
func (server *Server) ReachableCheck() bool {

	// 1. 设置过期时间并发送请求，2 秒足够了
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	// Head 方法只获取响应的 header，加快传输速度
	resp, err := client.Head(server.URL.String())

	// 2. 出错了就设置为 false
	if err != nil || resp.StatusCode != http.StatusOK {
		server.SetAlive(false)
		return false
	}

	// 3. 请求成功
	server.SetAlive(true)
	return true
}

func (server *Server) getRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(RetriesKey).(int); ok {
		return retry
	}
	return 0
}
