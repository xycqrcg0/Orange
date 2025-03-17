package main

import (
	"log"
	"net"
)

func dealCommand() {

}

func main() {
	//申请连接
	conn, err := net.Dial("tcp", "127.0.0.1:9979")
	if err != nil {
		log.Println("connection failure,", err)
		return
	}

	defer conn.Close()

	msg1 := GenerateMsg("message from client")
	_, err = conn.Write(msg1)
	if err != nil {
		log.Println("信息发送失败")
		return
		//continue
	}
	msg2 := GenerateMsg("message from client")
	_, err = conn.Write(msg2)
	if err != nil {
		log.Println("信息发送失败")
		return
		//continue
	}

}
