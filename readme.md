
## 说明

编写简单的负载均衡器，功能：

- 轮询分发
- 失败重试
- 后端可用性检测

## 视频链接

- [013. 负载均衡器第一部分：从零开始构建负载均衡器](https://learnku.com/courses/go-video/2022/building-load-balancers-from-scratch/11667)
- [014. 负载均衡器第二部分：可用服务器监测](https://learnku.com/courses/go-video/2022/available-server-monitoring/11668)

## 运行代码

```
go run .
```

## 后端服务

启动后端服务：

```
go run backend/myapp.go -port=6000 > /dev/null 2>&1 &
go run backend/myapp.go -port=6001 > /dev/null 2>&1 &
go run backend/myapp.go -port=6002 > /dev/null 2>&1 &
go run backend/myapp.go -port=6003 > /dev/null 2>&1 &
go run backend/myapp.go -port=6004 > /dev/null 2>&1 &
```

停止所有后端服务：

```
kill -9 $(lsof -t -i:6000,6001,6002,6003,6004 -sTCP:LISTEN)
```

停止单个后端：

```
kill -9 $(lsof -t -i:6004 -sTCP:LISTEN)
```