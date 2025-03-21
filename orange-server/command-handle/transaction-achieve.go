package command_handle

import (
	"bufio"
	"io"
	"log"
	"net"
	"orange-server/models"
	"orange-server/utils"
)

func createDBCopy() (dbCp *Base) {
	//把数据库遍历一遍，遍历的时候锁一下
	DB.Mtx.Lock()
	defer DB.Mtx.Unlock()
	dbCp = &Base{Data: make([]*ONode, 1024), Length: 1024, Sum: 0}
	for index, data := range DB.Data {
		if data != nil {
			vNode := &ONode{Next: nil}
			n := vNode
			for data != nil {
				node := &ONode{}
				node.Key = data.Key
				var value interface{}
				//蜜汁操作
				if dv, ok := data.Value.(*models.SDS); ok {
					v := *dv
					value = &v
				}
				if dv, ok := data.Value.(*OListNode); ok {
					v := *dv
					value = &v
				}
				if dv, ok := data.Value.(*OHash); ok {
					v := *dv
					value = &v
				}
				if dv, ok := data.Value.(*OSet); ok {
					v := *dv
					value = &v
				}
				node.Value = value

				n.Next = node
				n = n.Next

				data = data.Next
			}
			dbCp.Data[index] = vNode.Next
		}
	}
	return dbCp
}

func Transaction(conn net.Conn) {
	//创建一个当前数据库的快照(copy)
	dbCp := createDBCopy()

	//前置
	msg := utils.GenerateMsg("OK")
	conn.Write(msg)
	//开始记录命令
	//缓冲区
	buf := make([]byte, 512)
	wbuf := make([]string, 0)

	reader := bufio.NewReader(conn)
	for {
		msg = utils.GenerateMsg("Illegal Input")
		//写入缓冲区
		_, err := reader.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("读取失败,", err)
			break
		}
		//这里每次收到的命令都只有一条
		commandByte := buf[:]
		//commandByte = bytes.TrimRightFunc(commandByte, func(r rune) bool {
		//	return r == 0 // 去掉末尾的零字节
		//})
		_, _, com := utils.ParseMsg(commandByte)
		command := com[0]

		if command == "reset" {
			//退出事务，其他啥也不用干了
			msg := utils.GenerateMsg("OK, now transaction is exited")
			conn.Write(msg)
			break
		}

		if command == "commit" {
			//提交事务
			DB.Mtx.Lock()
			for _, c := range wbuf {
				DB.WriteAssign(c)
			}
			DB.Mtx.Unlock()
			msg := utils.GenerateMsg("OK, now transaction has been commited")
			conn.Write(msg)
			break
		}

		rmsg := dbCp.ReadAssign(command)
		if len(rmsg) != 0 {
			msg = rmsg
		}

		wmsg, ok := dbCp.WriteAssign(command)
		if ok {
			//命令写入wBuf
			wbuf = append(wbuf, command)
			msg = utils.GenerateMsg("ADDED")
		} else if len(wmsg) != 0 {
			msg = wmsg
		}

		conn.Write(msg)
	}
}
