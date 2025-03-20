package main

import (
	"log"
	"orange-server/command-handle"
)

func Init() {
	//ODB文件里的数据先写入
	if err := command_handle.ReadODB(); err != nil {
		log.Println("当前不存在.odb文件")
	}

	command_handle.Stop = make(chan bool)

	//初始默认值
	go command_handle.Save(5, 5)
}
