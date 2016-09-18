package main

import (
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	go serverProxyConn()
	go serverMainConn()
	go serverTransConn()
	time.Sleep(999999 * time.Hour)
}

func serverProxyConn() {
	l, err := net.Listen("tcp", ":7777")
	if nil != err {
		log.Fatalln("listen error", err)
		return
	}

	for {
		conn, err := l.Accept()
		if nil != err {
			log.Fatalln("accept error", err)
			break
		}

		go handlerProxyConn(conn)
	}
}

func handlerProxyConn(conn net.Conn) {

	// defer conn.Close()

	//GET /translate/releases/twsfe_w_20160912_RC07/r/js/desktop_module_main.js HTTP/1.1
	//Host: translate.google.cn
	//User-Agent: Mozilla/5.0 (X11; Linux i686; rv:45.0) Gecko/20100101 Firefox/45.0
	//Accept: */*
	//Accept-Language: en-US,en;q=0.5
	//Accept-Encoding: gzip, deflate
	//Referer: http://translate.google.cn/
	//Cookie: _ga=GA1.3.809376144.1441551016; NID=80=bZizxYb5VsCvLQMZwXS-CX7qRSVTdNKOfTk9C3WDMBSzEXqsvrS4KLNxYETmJtJIM9a_8uDi2xQ5nceuLzdQIEoZY7B5pZZorcbGrtFmkhz-k8OYLcX4lRNSilQHVafc
	//Connection: close

	ch := make(chan ProxyMsg)
	go transProxy(ch, conn)

	for {
		buf := make([]byte, 2048)
		leng, err := conn.Read(buf)
		msg := ProxyMsg{Leng: leng, Err: err, Data: buf}
		ch <- msg
	}

}

func transProxy(ch chan ProxyMsg, proxyConn net.Conn) {

	init := false
	id := makeRandomId()
	proxywork := ProxyWork{Id: id}

	for {
		msg := <-ch

		if init == false {

			//解析http头，得到请求的域名
			host := getHosts(msg.Data)

			//找到域名对应的mainConn, 让客户端发一个请求上来
			notifyClient(host, id)

			// 等待客户端的请求，转发数据
			for {
				transConn := getTransConn(id)
				if nil != transConn {
					proxywork.ProxyConn = proxyConn
					proxywork.TransConn = transConn
					init = true
					break
				}

				time.Sleep(10 * time.Millisecond)
			}
		}

		if msg.Err != nil {
			// 连接关闭
			proxywork.ProxyConn.Close()
			proxywork.TransConn.Close()
			break
		} else {
			proxywork.TransConn.Write(msg.Data)
		}

	}
}

func notifyClient(host, id string) {

}

func getTransConn(id string) net.Conn {
	return nil
}

func getHosts(bys []byte) string {

	str := string(bys)
	strs := strings.Split(str, "\n")

	for _, v := range strs {
		v = strings.TrimSpace(v)
		v = strings.Replace(v, " ", "", -1)

		if strings.HasPrefix(v, "Host:") {
			return strings.TrimLeft(v, "Host:")
		}
	}
	return ""
}

func makeRandomId() string {
	return ""
}

// 监听主线程
func serverMainConn() {
	l, err := net.Listen("tcp", ":8888")
	if nil != err {
		log.Fatalln("listen error", err)
		return
	}

	for {
		conn, err := l.Accept()
		if nil != err {
			log.Fatalln("accept error", err)
			break
		}

		go handlerMainConn(conn)
	}
}

func handlerMainConn(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 512)
		_, err := conn.Read(buf)
		if nil != err {
			log.Fatalln("read error", err)
			break
		}

		// 首次连接时，提供用户名密码
		// 得到请求域名和用户名的对应关系
		// 绑定用户名与Conn的关系

	}
}

// 监听数据传输端口
func serverTransConn() {
	l, err := net.Listen("tcp", ":9999")
	if nil != err {
		log.Fatalln("listen error", err)
		return
	}

	for {
		conn, err := l.Accept()
		if nil != err {
			log.Fatalln("accept error", err)
			break
		}

		go handlerTransConn(conn)
	}
}

func handlerTransConn(conn net.Conn) {

}
