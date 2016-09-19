package main

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/alexxuyao/kiwiproxy/common"
)

func Test_conn(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8888")
	//conn.Write([]byte("hello , what is the fuck"))

	for i := 0; i < 10; i++ {

		if nil != err {
			fmt.Println("not connection ")
			return
		}

		handshake := common.HandShakeMsg{Username: "alexxu", Password: "this is my password", Domains: []string{"127.0.0.1", "120.1.1.1" + strconv.Itoa(i)}}

		cnt, err := common.PacketMsg(handshake, common.MSG_TYPE_HAND_SHAKE)
		if nil != err {
			fmt.Println("packet message error:", err)
		}

		_, err = conn.Write(cnt)
		if nil != err {
			fmt.Println("write message error", err)
		}

		time.Sleep(20 * time.Minute)
	}
	conn.Close()

}
