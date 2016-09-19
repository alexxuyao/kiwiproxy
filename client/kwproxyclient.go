package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/alexxuyao/kiwiproxy/common"
)

func main() {
	client := KwProxyClient{}
	client.Start()
}

type KwProxyClient struct {
	config common.ClientConfig
}

// 读取配置
func (client *KwProxyClient) readConfig() {

	client.config.Domains = make([]common.DomainPair, 20)

	filepath := "./config.json"
	r, err := os.Open(filepath)
	defer r.Close()

	if nil != err {
		log.Println("open config file error,", err)
		return
	}

	bs, _ := ioutil.ReadAll(r)
	json.Unmarshal(bs, &client.config)

	ret, _ := json.Marshal(client.config)
	log.Println(string(ret))
}

func (client *KwProxyClient) Start() {

	// 读取配置
	client.readConfig()

	return

	// 起主连接
	conn, err := net.Dial("tcp", "localhost:8888")

	if nil != err {
		log.Println("not connection ")
		return
	}

	// 握手
	domains := make([]string, len(client.config.Domains))
	for _, v := range client.config.Domains {
		domains = append(domains, v.Remote)
	}

	handshake := common.HandShakeMsg{Username: client.config.Username, Password: client.config.Password, Domains: domains}

	cnt, err := common.PacketMsg(handshake, common.MSG_TYPE_HAND_SHAKE)
	if nil != err {
		log.Println("packet message error:", err)
	}

	_, err = conn.Write(cnt)
	if nil != err {
		log.Println("write message error", err)
	}

	// 接收消息
	client.handlerConn(conn)
}

func (client *KwProxyClient) handlerConn(conn net.Conn) {
	for {

		_, msgType, data, err := common.ReadMsg(conn)

		if nil != err {
			log.Println("read msg error,", err)
			break
		} else {
			go client.handlerMainMsg(msgType, data, conn)
		}

	}
}

// 处理消息
func (client *KwProxyClient) handlerMainMsg(msgType uint16, data []byte, conn net.Conn) {

	if common.MSG_TYPE_CREATE_TRANS_CONN == msgType {

		trans := common.CreateTransConnMsg{}
		err := json.Unmarshal(data, &trans)

		if nil != err {
			log.Println("unmarshal error,", err)
		}

		// 起一个传输连接

		transConn, err := net.Dial("tcp", "localhost:9999")

		if nil != err {
			log.Println("not connection ")
			return
		}

		// 起一个本地连接，连接本地http服务器
		localAddr := client.getLocalAddr(trans.Domain)
		localConn, err := net.Dial("tcp", localAddr)

		if nil != err {
			log.Println("not connection ")
			transConn.Close()
			return
		}

		// 消息原样返回
		cnt, err := common.PacketMsg(trans, common.MSG_TYPE_CREATE_TRANS_CONN)
		if nil != err {
			log.Println("packet message error,", err)
		}

		transConn.Write(cnt)

		go client.handleReadLeftToRight(transConn, localConn)
		go client.handleReadLeftToRight(localConn, transConn)

	}
}

// 从左边的连接读取数据，写到右边
func (client *KwProxyClient) handleReadLeftToRight(transConn, localConn net.Conn) {
	for {
		buf := make([]byte, 2048)
		leng, err := transConn.Read(buf)

		if err != nil {
			log.Println("read error, ", err)

		}

		_, err = localConn.Write(buf[0:leng])

		if err != nil {
			log.Println("read error, ", err)

		}

	}
}

func (client *KwProxyClient) getLocalAddr(remoteAddr string) string {
	return ""
}
