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

	_, err = conn.Write([]byte("message from client"))
	if err != nil {
		log.Println("信息发送失败")
		return
		//continue
	}

}
