package command_handle

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"orange-server/utils"
)

func Transaction(conn net.Conn) {
	//前置
	msg := utils.GenerateMsg("OK")
	conn.Write(msg)
	//开始记录命令
	//缓冲区
	rbuf := make([]byte, 512)
	wbuf := make([]byte, 0)

	reader := bufio.NewReader(conn)
	for {
		//写入缓冲区
		_, err := reader.Read(rbuf[:])
		if err != nil {
			if err == io.EOF {
				//if m > 0 {
				//	da := string(buf[:m])
				//	fmt.Println(da)
				//}
				break
			}
			log.Println("读取失败,", err)
			break
		}
		//这里每次收到的命令都只有一条
		commandByte := rbuf[:]
		commandByte = bytes.TrimRightFunc(commandByte, func(r rune) bool {
			return r == 0 // 去掉末尾的零字节
		})
		_, _, command := utils.ParseMsg(commandByte)

		if command[0] == "reset" {
			//退出事务
			msg := utils.GenerateMsg("OK, now transaction is exited")
			conn.Write(msg)
			break

		}
		if command[0] == "commit" {
			//丢给commandAssign解决,但是它会向客户端发送响应（嘶）
			_, _, commands := utils.ParseMsg(wbuf)
			CommandsAssign(conn, commands)
			//提交事务
			msg := utils.GenerateMsg("OK, now transaction has been commited")
			conn.Write(msg)
			break
		}

		f := 0
		for _, pattern := range patterns {
			if pattern.MatchString(command[0]) {
				wbuf = append(wbuf, commandByte...)
				msg := utils.GenerateMsg("ADDED")
				f = 1
				conn.Write(msg)
				break
			}
		}
		if f == 0 {
			//没匹配到，那么这个命令就是不合法的
			mmsg := utils.GenerateMsg("Illegal Input")
			conn.Write(mmsg)
		}
	}
}
