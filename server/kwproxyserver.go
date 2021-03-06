package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/alexxuyao/kiwiproxy/common"
	"github.com/streamrail/concurrent-map"
)

func main() {
	app := ProxyApp{}
	app.Start()
}

// 主服务
type ProxyApp struct {
	domainMap     cmap.ConcurrentMap // 域名与用户名的关系 map[domain]username
	mainConnMap   cmap.ConcurrentMap // 用户名与主连接的关系 map[username]net.Conn
	handlShakeMap cmap.ConcurrentMap // 用户名与HandShake的关系 map[username]*common.HandShakeMsg
	proxyworkMap  cmap.ConcurrentMap // 消息ID与proxywork的关系 map[msgId]*ProxyWork
	genIdLock     sync.Mutex         // 生成消息ID时会使用的锁
	msgId         uint32             // 消息ID
	config        common.ServerConfig
}

// 开始监听
func (app *ProxyApp) Start() {

	app.domainMap = cmap.New()
	app.mainConnMap = cmap.New()
	app.handlShakeMap = cmap.New()
	app.proxyworkMap = cmap.New()

	// 读取配置
	app.readConfig()

	// 启动监听
	go app.serverProxyConn()

	// 启动监听传输连接
	go app.serverTransConn()

	// 启动监听主连接
	app.serverMainConn()
}

// 读取配置
func (app *ProxyApp) readConfig() {

	filepath := "./config.json"
	r, err := os.Open(filepath)
	defer r.Close()

	if nil != err {
		log.Println("open config file error,", err)
		return
	}

	bs, _ := ioutil.ReadAll(r)
	json.Unmarshal(bs, &app.config)

	ret, _ := json.Marshal(app.config)
	log.Println(string(ret))
}

// 监听某一个端口
func listenServer(addr string, handler func(net.Conn)) {

	l, err := net.Listen("tcp", addr)

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

		go handler(conn)
	}
}

//
func (app *ProxyApp) serverProxyConn() {
	listenServer(app.config.ProxyConnServer, app.handlerProxyConn)
}

func (app *ProxyApp) handlerProxyConn(conn net.Conn) {

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

	log.Println("handlerProxyConn accept conn", conn)

	ch := make(chan ProxyMsg)
	go app.transProxy(ch, conn)

	for {
		buf := make([]byte, 2048)
		leng, err := conn.Read(buf)
		msg := ProxyMsg{Leng: leng, Err: err, Data: buf[0:leng]}
		ch <- msg
	}

}

// 数据传输
func (app *ProxyApp) transProxy(ch chan ProxyMsg, proxyConn net.Conn) {

	init := false
	id := app.genMsgId()
	tch := make(chan uint32)
	proxywork := &ProxyWork{Id: id, Chan4TransConn: tch, ProxyConn: proxyConn}
	app.proxyworkMap.Set(strconv.FormatUint(uint64(id), 10), proxywork)

	for {
		msg := <-ch

		if msg.Err != nil {
			log.Println("get error:", msg.Err)
			break
		}

		log.Println("transProxy, get data, ", string(msg.Data))

		if init == false {

			//解析http头，得到请求的域名
			domain := app.getRequestDomain(msg.Data)

			log.Println("get request domain in transProxy, ", domain)

			//找到域名对应的mainConn, 让客户端发一个请求上来
			err := app.notifyClient(domain, id)

			if err != nil {
				break
			}

			// 等待客户端的请求，转发数据
			cMsgId := <-proxywork.Chan4TransConn

			go common.ReadLeftToRight(proxywork.TransConn, proxywork.ProxyConn)

			log.Println("transProxy, init finish, client MsgId", cMsgId)

			init = true

		}

		log.Println("write conn ,local:", proxywork.TransConn.LocalAddr(), ", remote:", proxywork.TransConn.RemoteAddr())

		wlen, err := proxywork.TransConn.Write(msg.Data)

		if nil != err {
			log.Println("write to trans conn error:", err)
		}

		log.Println("transProxy, conn write,len : ", wlen, ", msgId:", id, ", msg:", string(msg.Data))
	}

	app.releaseProxyWork(proxywork)
}

func (app *ProxyApp) releaseProxyWork(proxywork *ProxyWork) {

	// 连接关闭
	proxywork.ProxyConn.Close()

	if nil != proxywork.TransConn {
		proxywork.TransConn.Close()
	}

	app.proxyworkMap.Remove(strconv.FormatUint(uint64(proxywork.Id), 10))

	log.Println("transProxy, conn close, msgId:", proxywork.Id)
}

func (app *ProxyApp) notifyClient(domain string, id uint32) error {

	if tmp, ok := app.domainMap.Get(domain); ok {
		username := tmp.(string)

		if tmp, ok := app.mainConnMap.Get(username); ok {
			conn := tmp.(net.Conn)

			createMsg := common.CreateTransConnMsg{MsgId: id, Domain: domain}
			ret, err := common.PacketMsg(createMsg, common.MSG_TYPE_CREATE_TRANS_CONN)

			if nil == err {
				conn.Write(ret)
				return nil
			}

		} else {
			log.Println("can not found conn for username, ", username)
		}

	} else {
		log.Println("can not found username for domain, ", domain)
	}

	return errors.New("find client error for " + domain + ", msgId:" + strconv.FormatUint(uint64(id), 10))

}

// 从http头中解析出host
func (app *ProxyApp) getRequestDomain(bys []byte) string {

	str := string(bys)
	strs := strings.Split(str, "\n")

	for _, v := range strs {
		v = strings.TrimSpace(v)
		v = strings.Replace(v, " ", "", -1)

		if strings.HasPrefix(v, "Host:") {
			domain := strings.TrimLeft(v, "Host:")
			if strings.Contains(domain, ":") {
				domain = strings.Split(domain, ":")[0]
			}
			return domain
		}
	}

	return ""
}

func (app *ProxyApp) genMsgId() uint32 {
	app.genIdLock.Lock()
	defer app.genIdLock.Unlock()

	// uint32最大值
	if app.msgId == 4294967295 {
		app.msgId = 0
	}

	app.msgId += 1

	return app.msgId
}

// 监听主线程
func (app *ProxyApp) serverMainConn() {
	listenServer(app.config.MainConnServer, app.handlerMainConn)
}

func (app *ProxyApp) handlerMainConn(conn net.Conn) {
	defer conn.Close()

	var handshake *common.HandShakeMsg

	for {

		_, msgType, data, err := common.ReadMsg(conn)

		if nil != err {
			log.Println("read msg error,", err)
			break
		}

		if common.MSG_TYPE_HAND_SHAKE == msgType {
			handshake = app.handlerHandShakeMsg(msgType, data, conn)
		} else {
			go app.handlerMainMsg(msgType, data, conn)
		}

	}

	app.releaseResource(handshake)
}

// 客户端断开连接开，释放资源
func (app *ProxyApp) releaseResource(handshake *common.HandShakeMsg) {
	log.Println("release resource for ,", handshake.Username)

	// TODO 释放资源, 那几个map

}

// 处理消息
func (app *ProxyApp) handlerMainMsg(msgType uint16, data []byte, conn net.Conn) {

}

// 处理消息
func (app *ProxyApp) handlerHandShakeMsg(msgType uint16, data []byte, conn net.Conn) *common.HandShakeMsg {

	handshake := common.HandShakeMsg{}
	err := json.Unmarshal(data, &handshake)
	if nil != err {
		log.Println("unmarshal error,", err)
	}

	// 首次连接时，校验用户名密码

	// 得到请求域名和用户名的对应关系
	for _, v := range handshake.Domains {
		app.domainMap.Set(v, handshake.Username)
	}

	// 绑定用户名与Conn的关系
	app.mainConnMap.Set(handshake.Username, conn)

	// 绑定用户名与HandShakeMsg的关系
	app.handlShakeMap.Set(handshake.Username, &handshake)

	log.Println("do handshake finish, username", handshake.Username, ", domains:", handshake.Domains)

	return &handshake
}

// 监听数据传输端口
func (app *ProxyApp) serverTransConn() {
	listenServer(app.config.TransConnServer, app.handlerTransConn)
}

func (app *ProxyApp) handlerTransConn(conn net.Conn) {
	_, msgType, data, err := common.ReadMsg(conn)

	if nil != err {
		log.Println("handlerTransConn error, ", err)
	} else {
		if common.MSG_TYPE_CREATE_TRANS_CONN == msgType {

			msg := common.CreateTransConnMsg{}
			err := json.Unmarshal(data, &msg)

			if nil != err {
				log.Println("unmarshal error,", err)
			}

			if tmp, ok := app.proxyworkMap.Get(strconv.FormatUint(uint64(msg.MsgId), 10)); ok {
				proxywork := tmp.(*ProxyWork)
				proxywork.TransConn = conn
				proxywork.Chan4TransConn <- msg.MsgId

				log.Println("attach proxywork trans conn, msgId:", msg.MsgId, ", transConn is :", proxywork.TransConn)
			}
		}
	}
}
