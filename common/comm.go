package common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
)

const (
	MSG_TYPE_HAND_SHAKE        uint16 = 1 // 消息类型：握手
	MSG_TYPE_CREATE_TRANS_CONN uint16 = 2 // 消息类型：服务端通知客户端建立传输连接
)

type DomainPair struct {
	Remote string `json:"remote"` // 远程域名,这个不需要端口，如www.baidu.com
	Local  string `json:"local"`  // 本地服务器域名，可加可不加端口，如127.0.0.1:8080
}

// 客户端配置
type ClientConfig struct {
	Username        string       `json:"username"`        // 用户名
	Password        string       `json:"password"`        // 密码
	MainConnServer  string       `json:"mainConnServer"`  // 主连接服务器
	TransConnServer string       `json:"transConnServer"` // 传输连接服务器
	Domains         []DomainPair `json:"domains"`         // 绑定的域名
}

// 服务端配置
type ServerConfig struct {
	MainConnServer  string `json:"mainConnServer"`  // 主连接服务器
	ProxyConnServer string `json:"proxyConnServer"` // http监听端口
	TransConnServer string `json:"transConnServer"` // 传输连接服务器
}

// 客户端发起握手
type HandShakeMsg struct {
	Username string   `json:"username"` // 用户名
	Password string   `json:"password"` // 密码
	Domains  []string `json:"domains"`  // 绑定的域名
}

// 服务端让客户端建立传输连接
type CreateTransConnMsg struct {
	MsgId  uint32 `json:"msgId"`  // 代理作业的唯一ID
	Domain string `json:"domain"` // 传输的domain
}

// 包装消息
func PacketMsg(obj interface{}, msgType uint16) ([]byte, error) {
	cnt, err := json.Marshal(obj)

	if nil != err {
		return nil, err
	}

	dlen := uint32(len(cnt) + 2)

	// 消息长度
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, dlen)

	// 消息类型
	msgTypeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(msgTypeBuf, msgType)

	ret := make([]byte, len(cnt)+6)
	copy(ret, buf)
	copy(ret[4:6], msgTypeBuf)
	copy(ret[6:], cnt)

	return ret, nil
}

// 读消息，返回数据长度，消息类型，消息体，错误
func ReadMsg(conn net.Conn) (uint32, uint16, []byte, error) {
	// 读取消息长度,消息长度使用uint32表示，4个字节
	buf := make([]byte, 4)
	blen, err := conn.Read(buf)
	if nil != err {
		log.Println("read len error:", err)
		return 0, 0, nil, err
	}

	if blen != 4 {
		// 读取长度错误
		return 0, 0, nil, errors.New("read len error")
	}

	// 得到消息长度，开始读取数据
	datalen := binary.BigEndian.Uint32(buf)
	data := make([]byte, datalen)
	var readlen uint32 = 0

	log.Println("datalen is :", datalen)

	for {
		buflen := 1024

		if datalen-readlen < 1024 {
			buflen = int(datalen - readlen)
		}

		dbuf := make([]byte, buflen)
		dlen, err := conn.Read(dbuf)

		if nil != err {
			// 读取数据错误
			log.Println("read data error:", err)
			return 0, 0, nil, err
		}

		log.Println("readlen is :", readlen, ", dlen is :", dlen, ",dbuf is :", dbuf)
		copy(data[readlen:readlen+uint32(dlen)], dbuf)

		readlen += uint32(dlen)

		if readlen >= datalen {
			break
		}
	}

	log.Println(string(data[2:]))

	return datalen, binary.BigEndian.Uint16(data[0:2]), data[2:], nil
}

// 从左边的连接读取数据，写到右边
func ReadLeftToRight(leftConn, rightConn net.Conn) {
	for {

		log.Println("read left conn ,local:", leftConn.LocalAddr(), ", remote:", leftConn.RemoteAddr())

		buf := make([]byte, 1024)
		leng, err := leftConn.Read(buf)

		if err != nil {
			log.Println("read left conn error, ", err, ",local:", leftConn.LocalAddr(), ", remote:", leftConn.RemoteAddr())
			break
		}

		log.Println("read content from left,", string(buf[0:leng]))

		_, err = rightConn.Write(buf[0:leng])

		if err != nil {
			log.Println("write right conn error, ", err, ",local:", rightConn.LocalAddr(), ", remote:", rightConn.RemoteAddr())
			break
		}

	}
}
