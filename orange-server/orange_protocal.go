package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

func GenerateMsg(body string) []byte {
	n := len(body)
	var nByte [4]byte
	nByte[3] = byte(n)
	nByte[2] = byte(n >> 8)
	nByte[1] = byte(n >> 16)
	nByte[0] = byte(n >> 32)
	msg := make([]byte, 0)
	msg = append(msg, 0x99, 0x79, nByte[0], nByte[1], nByte[2], nByte[3])
	msg = append(msg, []byte(body)...)
	return msg
}

// ParseMsg 返回值里的n表示有几条内容，p表示最后一条内容是否完整，不完整则p为偏移量，commands为内容
func ParseMsg(msg []byte) (n int, p int, contents []string) {
	n, p = 0, 0
	//point记录当前读到了msg的哪个位置
	point := 0
	contents = make([]string, 0)

	for {
		l := len(msg[point:])
		if l < 7 || msg[point] != 0x99 || msg[point+1] != 0x79 {
			break
		}

		point += 2
		nByte := bytes.NewBuffer(msg[point : point+4])
		point += 4
		var length32 int32
		var length int
		err := binary.Read(nByte, binary.BigEndian, &length32)
		if err != nil {
			log.Println(err)
		}
		length = int(length32)

		if l-6 < length {
			//该条信息是最后一条，而且不完整
			p = l - 6
			contents = append(contents, string(msg[point:]))
		} else {
			contents = append(contents, string(msg[point:point+length]))
		}
		point += length
		n++
	}

	return n, p, contents
}
