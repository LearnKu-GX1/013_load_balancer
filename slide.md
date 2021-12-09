---
theme: gaia
_class: lead
paginate: true
backgroundColor: #fff
marp: true
slide_tool: https://github.com/marp-team/marp-core/tree/main/themes
---

# **什么是负载均衡**

Nginx 常用的几种算法

---

![](https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2020/3/29/17125abbccba1a78~tplv-t2oaga2asx-watermark.awebp)

---

#### 1. 轮询（round-robin）

平均分配，如某台服务器不可用，能自动剔除。

```
http {
    upstream myapp1 {
        server 192.168.0.11;
        server 192.168.0.12;
        server 192.168.0.13;
    }
    server {
        listen 80;
        location / {
            proxy_pass http://myapp1;
        }
    }
}
```

---

#### 2. 加权（weight）

指定轮询几率，weight 和访问比率成正比，用于后端服务器性能不均的情况。

```
upstream backserver {
    server 192.168.0.11 weight=2;
    server 192.168.0.12 weight=6;
    server 192.168.0.13 weight=2;
}
```

---

#### 3. ip_hash

根据 IP 选择服务器，某个 IP 永远访问固定的某台服务器：

```
upstream myapp1 {
    ip_hash
    server 192.168.0.11;
    server 192.168.0.12;
    server 192.168.0.13;
}
```

---

## 我们的负载均衡器

功能：

- 轮询模式
- 可用性检测
- 线程安全


