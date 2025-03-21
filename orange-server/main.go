package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	commandhandle "orange-server/command-handle"
	"orange-server/utils"
)

func handle(conn net.Conn) {
	defer conn.Close()
	//缓冲区
	buf := make([]byte, 1024)
	reader := bufio.NewReader(conn)
	//p记录偏移量
	p := 0
	for {
		//写入缓冲区
		m, err := reader.Read(buf[p:])
		if err != nil {
			if err == io.EOF {
				if m > 0 {
					da := string(buf[:m])
					fmt.Println(da)
				}
				break
			}
			log.Println("读取失败,", err)
			break
		}
		n, point, commands := utils.ParseMsg(buf[p:])
		if point != 0 {
			//有不全信息
			p = point
			copy([]byte(commands[n-1]), buf)
			//不全信息就先不处理
			commandhandle.DB.CommandsAssign(conn, commands[:n-1])
			continue
		}
		//p置0
		p = 0
		commandhandle.DB.CommandsAssign(conn, commands)
	}
}

func main() {
	Init()

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
