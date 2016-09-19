package main

import "net"

type ProxyMsg struct {
	Leng int
	Err  error
	Data []byte
}

// 表示一次代理作业
type ProxyWork struct {
	Id             uint32      // 唯一ID
	ProxyConn      net.Conn    // 用户访问链接的连接
	TransConn      net.Conn    // 客户端对服务端的连接
	Chan4TransConn chan uint32 // 客户端传输连接成功后，用来告诉Proxy协和
}
