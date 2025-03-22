package command_handle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"orange-server/models"
	"os"
)

// Record 记录当前改动了几个数据,为save作参考
var Record int64 = 0

// SAVEFlag 标识是否有SAVE进程在执行
var SAVEFlag int64 = 0

// SaveF 标识自动触发是否在运行
var SaveF int64 = 0

// Stop 控制后台的save操作
var Stop chan bool

var (
	sds  byte = 0x00
	list byte = 0x01
	hash byte = 0x02
	set  byte = 0x07

	symbol = []byte("odb")

	start byte = 0xaa
	end   byte = 0xee

	odbFilePath = "./orange.odb"
)

func intToByte(a int) (b []byte) {
	b = make([]byte, 4)
	b[3] = byte(a)
	b[2] = byte(a >> 8)
	b[1] = byte(a >> 16)
	b[0] = byte(a >> 24)
	return
}

func byteToInt(b []byte) (a int, err error) {
	nByte := bytes.NewBuffer(b)
	var length32 int32
	err = binary.Read(nByte, binary.BigEndian, &length32)
	if err != nil {
		log.Println(err)
	}
	return int(length32), nil
}

func writeSDSIn(a *models.SDS) (b []byte) {
	b = make([]byte, 0)
	b = append(b, intToByte(a.Length)...)
	b = append(b, intToByte(a.Alloc)...)
	b = append(b, a.Buf[:a.Length]...)
	return
}

func GetSDS(l int, a int, s []byte) *models.SDS {
	buf := make([]byte, a)
	copy(buf, s)
	return &models.SDS{
		Length: l,
		Alloc:  a,
		Buf:    buf,
	}
}

func WriteODB() error {
	writeBuf := make([]byte, 0)

	writeBuf = append(writeBuf, symbol...)
	writeBuf = append(writeBuf, intToByte(DB.Sum)...)
	writeBuf = append(writeBuf, intToByte(DB.Length)...)
	for index, v := range DB.Data {
		if v != nil {
			//注意v也可能是链表···
			//index
			writeBuf = append(writeBuf, intToByte(index)...)

			for v != nil {
				//key
				writeBuf = append(writeBuf, writeSDSIn(&v.Key)...)
				//value

				if value, ok := v.Value.(*models.SDS); ok {
					writeBuf = append(writeBuf, sds)

					writeBuf = append(writeBuf, writeSDSIn(value)...)

				} else if value, ok := v.Value.(*OListNode); ok {
					writeBuf = append(writeBuf, list)

					//writeBuf = append(writeBuf, start)
					for value != nil {
						writeBuf = append(writeBuf, writeSDSIn(value.Content)...)
						//之间空一下
						if value.Right == nil {
							writeBuf = append(writeBuf, end)
							break
						}
						writeBuf = append(writeBuf, 0x00)
						value = value.Right
					}

				} else if value, ok := v.Value.(*OHash); ok {
					writeBuf = append(writeBuf, hash)

					writeBuf = append(writeBuf, intToByte(value.Sum)...)
					writeBuf = append(writeBuf, intToByte(value.Length)...)

					for i, p := range value.Value {
						if p != nil {
							//index
							writeBuf = append(writeBuf, intToByte(i)...)
							q := p
							for q != nil {
								//field
								writeBuf = append(writeBuf, writeSDSIn(&q.Field)...)
								//value
								writeBuf = append(writeBuf, writeSDSIn(q.Value)...)

								if q.Next != nil {
									writeBuf = append(writeBuf, start)
								}
								q = q.Next
							}
							writeBuf = append(writeBuf, end)
						}
					}

				} else if value, ok := v.Value.(*OSet); ok {
					writeBuf = append(writeBuf, set)

					writeBuf = append(writeBuf, intToByte(value.Sum)...)
					writeBuf = append(writeBuf, intToByte(value.Length)...)

					for i, p := range value.Value {
						if p != nil {
							//index
							writeBuf = append(writeBuf, intToByte(i)...)
							q := p
							for q != nil {
								//value
								writeBuf = append(writeBuf, writeSDSIn(q.Value)...)

								if q.Next != nil {
									writeBuf = append(writeBuf, start)
								}
								q = q.Next
							}
							writeBuf = append(writeBuf, end)
						}
					}
				}

				if v.Next != nil {
					writeBuf = append(writeBuf, start)
					v = v.Next
				} else {
					writeBuf = append(writeBuf, end)
					break
				}
			}
		}
	}

	//还是不能先创建文件再构造writeBuf，可能会丢失数据

	file, err := os.Create(odbFilePath)
	if err != nil {
		fmt.Printf("创建文件时出错: %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(writeBuf)
	if err != nil {
		return err
	}

	return nil
}

func ReadODB() error {
	DB = &Base{
		Sum:    0,
		Length: 1024,
		Max:    0,
		Data:   make([]*ONode, 1024),
	}

	file, err := os.ReadFile(odbFilePath)
	if err != nil {
		return err
	}
	l := len(file)
	//偏移量
	p := 0
	if l < 7 || file[0] != byte('o') || file[1] != byte('d') || file[2] != byte('b') {
		return errors.New("odb文件格式不符")
	}
	p += 3

	if DB.Sum, err = byteToInt(file[p : p+4]); err != nil {
		return err
	}
	p += 4
	if DB.Length, err = byteToInt(file[p : p+4]); err != nil {
		return err
	}
	p += 4

	//如果Length不是1024，还要修改一下，扩容

	i := DB.Sum
	for i > 0 {
		index, _ := byteToInt(file[p : p+4])
		p += 4

		vDataNode := &ONode{}
		vd := vDataNode
		for {
			keyLength, _ := byteToInt(file[p : p+4])
			p += 4
			keyAlloc, _ := byteToInt(file[p : p+4])
			p += 4
			key := GetSDS(keyLength, keyAlloc, file[p:p+keyLength])
			p += keyLength

			var value interface{}
			p++
			switch file[p-1] {
			case sds:
				valueLength, _ := byteToInt(file[p : p+4])
				p += 4
				valueAlloc, _ := byteToInt(file[p : p+4])
				p += 4
				value = GetSDS(valueLength, valueAlloc, file[p:p+valueLength])
				p += valueLength

			case list:
				contentLength, _ := byteToInt(file[p : p+4])
				p += 4
				contentAlloc, _ := byteToInt(file[p : p+4])
				p += 4
				content := GetSDS(contentLength, contentAlloc, file[p:p+contentLength])
				p += contentLength

				ln := &OListNode{Content: content, Left: nil, Right: nil}
				value = ln

				for file[p] != end {
					p++
					nLength, _ := byteToInt(file[p : p+4])
					p += 4
					nAlloc, _ := byteToInt(file[p : p+4])
					p += 4
					n := GetSDS(nLength, nAlloc, file[p:p+nLength])
					p += nLength

					node := &OListNode{Content: n, Left: nil, Right: nil}
					ln.Right = node
					node.Left = ln

					ln = node
				}
				p++

			case hash:
				hashSum, _ := byteToInt(file[p : p+4])
				p += 4
				hashLength, _ := byteToInt(file[p : p+4])
				p += 4
				hashValue := make([]*OHashNode, hashLength)

				s := hashSum
				for s > 0 {
					inHashIndex, _ := byteToInt(file[p : p+4])
					p += 4

					fieldLength, _ := byteToInt(file[p : p+4])
					p += 4
					fieldAlloc, _ := byteToInt(file[p : p+4])
					p += 4
					field := GetSDS(fieldLength, fieldAlloc, file[p:p+fieldLength])
					p += fieldLength

					valueLength, _ := byteToInt(file[p : p+4])
					p += 4
					valueAlloc, _ := byteToInt(file[p : p+4])
					p += 4
					ivalue := GetSDS(valueLength, valueAlloc, file[p:p+valueLength])
					p += valueLength

					//得到一个值
					s--

					hashNode := &OHashNode{Field: *field, Value: ivalue, Next: nil}
					h := hashNode

					for file[p] != end {
						//是个链表
						p++

						cfieldLength, _ := byteToInt(file[p : p+4])
						p += 4
						cfieldAlloc, _ := byteToInt(file[p : p+4])
						p += 4
						cfield := GetSDS(cfieldLength, cfieldAlloc, file[p:p+cfieldLength])
						p += cfieldLength

						cvalueLength, _ := byteToInt(file[p : p+4])
						p += 4
						cvalueAlloc, _ := byteToInt(file[p : p+4])
						p += 4
						civalue := GetSDS(cvalueLength, cvalueAlloc, file[p:p+cvalueLength])
						p += cvalueLength

						chashNode := &OHashNode{Field: *cfield, Value: civalue, Next: nil}

						h.Next = chashNode
						h = h.Next

						//又得到一个值
						s--
					}
					p++

					hashValue[inHashIndex] = hashNode
				}
				value = &OHash{Length: hashLength, Sum: hashSum, Value: hashValue}

			case set:
				setSum, _ := byteToInt(file[p : p+4])
				p += 4
				setLength, _ := byteToInt(file[p : p+4])
				p += 4
				setValue := make([]*OSetNode, setLength)

				s := setSum
				for s > 0 {
					inSetIndex, _ := byteToInt(file[p : p+4])
					p += 4

					svalueLength, _ := byteToInt(file[p : p+4])
					p += 4
					svalueAlloc, _ := byteToInt(file[p : p+4])
					p += 4
					svalue := GetSDS(svalueLength, svalueAlloc, file[p:p+svalueLength])
					p += svalueLength

					setNode := &OSetNode{Value: svalue, Next: nil}
					se := setNode

					s--

					for file[p] != end {
						//是个链表
						p++

						ssvalueLength, _ := byteToInt(file[p : p+4])
						p += 4
						ssvalueAlloc, _ := byteToInt(file[p : p+4])
						p += 4
						ssvalue := GetSDS(ssvalueLength, ssvalueAlloc, file[p:p+ssvalueLength])
						p += ssvalueLength

						ssetNode := &OSetNode{Value: ssvalue, Next: nil}
						se.Next = ssetNode
						se = se.Next

						//又得到一个值
						s--
					}
					p++

					setValue[inSetIndex] = setNode
				}
				value = &OSet{Length: setLength, Sum: setSum, Value: setValue}

			}

			dataNode := &ONode{Key: *key, Value: value}
			vd.Next = dataNode
			vd = vd.Next
			i--

			if file[p] == end {
				p++
				break
			} else {
				p++
			}
		}

		DB.Data[index] = vDataNode.Next
	}
	return nil
}
