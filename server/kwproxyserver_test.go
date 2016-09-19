package main

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/alexxuyao/kiwiproxy/common"
)

func Test_getHosts(t *testing.T) {
	app := ProxyApp{}

	str := "hello world\r\nHost:www.baidu.com "
	by := []byte(str)
	fmt.Println(app.getRequestDomain(by))
}

func Test_byte(t *testing.T) {
	var mySlice = []byte{255, 255, 255, 255}
	data := binary.BigEndian.Uint32(mySlice)
	fmt.Println(data)

	//	var dlen uint32
	//	dlen = 0
	//	dlen += uint32(11)
	//	fmt.Println(dlen)

}

func Test_PacketMsg(t *testing.T) {
	ret, err := common.PacketMsg(common.CreateTransConnMsg{MsgId: uint32(50), Domain: "www.baidu.com"}, common.MSG_TYPE_CREATE_TRANS_CONN)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(ret[6:]))
}
