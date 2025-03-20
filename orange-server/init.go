package main

import (
	"bufio"
	"fmt"
	"log"
	"orange-server/command-handle"
	"os"
)

func Init() {
	//ODB文件里的数据先写入
	if err := command_handle.ReadODB(); err != nil {
		log.Println("当前不存在.odb文件")
	}

	command_handle.Stop = make(chan bool)

	command_handle.AOFBuf = make([]byte, 0)

	//初始默认值
	go command_handle.Save(5, 5)

	file, _ := os.Open("./orange.aof")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Bytes())
	}
	file.Close()

}
