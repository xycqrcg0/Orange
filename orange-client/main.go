package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	//申请连接
	conn, err := net.Dial("tcp", "127.0.0.1:9979")
	if err != nil {
		log.Println("connection failure,", err)
		return
	}

	defer conn.Close()

	readerIn := bufio.NewReader(os.Stdin)

	buf := make([]byte, 1024)
	readerOut := bufio.NewReader(conn)

	for {
		command, _ := readerIn.ReadString('\n')
		msg := GenerateMsg(command[:len(command)-1])
		conn.Write(msg)

		if command == "exit\n" {
			fmt.Println("ByeBye~")
			break
		}

		m, err := readerOut.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				if m > 0 {
					da := string(buf[:m])
					fmt.Println(da)
					fmt.Println("---数据库连接断开---")
				}
				break
			}
			fmt.Println("---!异常!---")
			break
		}
		_, _, contents := ParseMsg(buf[:])
		for _, content := range contents {
			fmt.Println(content)
		}
	}

}
