package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func handle(conn net.Conn) {
	defer conn.Close()
	//缓冲区
	buf := make([]byte, 1024)
	reader := bufio.NewReader(conn)
	for {
		//写入缓冲区
		n, err := reader.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				if n > 0 {
					data := string(buf[:n])
					fmt.Println(data)
				}
				break
			}
			log.Println("读取失败,", err)
			break
		}
		data := string(buf[:n])
		fmt.Println(data)
	}
}

func main() {
	//监听端口
	tcpSocket, err := net.Listen("tcp", "127.0.0.1:9979")
	if err != nil {
		log.Println("socket创建失败~")
		return
	}
	for {
		conn, err := tcpSocket.Accept()
		if err != nil {
			log.Println("accept failure,", err)
			continue
		}
		//开线程
		go handle(conn)
	}
}
