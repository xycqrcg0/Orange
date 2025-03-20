package command_handle

import (
	"log"
	"orange-server/data"
	"orange-server/models"
	"orange-server/utils"
	"os"
	"sync"
	"sync/atomic"
)

//嘶，原来aof重写读的不是原来的aof文件，而是直接遍历数据库啊···

// AOFStatus 标识AOF功能是否开启
var AOFStatus int64

var (
	aofFilePath = "./orange.aof"

	AOFSize int64 = 0

	// AOFFlag 标识当前是否有aof重写正在进行
	AOFFlag int64 = 0

	mtx sync.Mutex

	// AOFBuf 重写过程中新数据存放的缓冲区(init()里初始化)
	AOFBuf []byte
	// Bmtx 为AOFBuf准备的锁
	Bmtx sync.Mutex
)

// WriteInAOFBuf 保证并发安全
func WriteInAOFBuf(msg []byte) {
	Bmtx.Lock()
	AOFBuf = append(AOFBuf, msg...)
	Bmtx.Unlock()
}

// AOFReread 负责读取数据库，将数据转化为命令
func AOFReread() (read []byte) {

	//读的过程中database被修改了怎么办？从读开始，接下来接收到的写入命令都存起来，之后接到此read之后

	read = make([]byte, 0)
	for _, v := range data.Database.Data {
		for v != nil {
			key := string(v.Key.Buf[:v.Key.Length])

			if vv, ok := v.Value.(*models.SDS); ok {
				value := string(vv.Buf[:vv.Length])
				command := "set(" + key + "," + value + ")"
				msg := utils.GenerateMsg(command)
				read = append(read, msg...)

			} else if vv, ok := v.Value.(*OListNode); ok {
				for vv != nil {
					value := string(vv.Content.Buf[:vv.Content.Length])
					command := "addr(" + key + "," + value + ")"
					msg := utils.GenerateMsg(command)
					read = append(read, msg...)
					vv = vv.Right
				}

			} else if vv, ok := v.Value.(*OHash); ok {
				for _, vvf := range vv.Value {
					if vvf != nil {
						for vvf != nil {
							field := string(vvf.Field.Buf[:vvf.Field.Length])
							value := string(vvf.Value.Buf[:vvf.Value.Length])
							command := "hset(" + key + "," + field + "," + value + ")"
							msg := utils.GenerateMsg(command)
							read = append(read, msg...)
							vvf = vvf.Next
						}
					}
				}

			} else if vv, ok := v.Value.(*OSet); ok {
				for _, vvv := range vv.Value {
					if vvv != nil {
						value := string(vvv.Buf[:vvv.Length])
						command := "sadd(" + key + "," + value + ")"
						msg := utils.GenerateMsg(command)
						read = append(read, msg...)
					}
				}
			}
			v = v.Next
		}
	}
	return read
}

// AOF 向aof文件里添上新命令,顺便检查一下当前aof文件需不需要重写,AND这个操作应该是要加锁的
func AOF(msg []byte) {
	mtx.Lock()

	//检查aof文件是否存在
	file, err := os.OpenFile(aofFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666) //权限没想好怎么设置
	if err != nil {
		log.Println("aof文件打开/创建失败")
		//没想好这里怎么处理
		return
	}
	//写入
	log.Println(msg)
	file.Write(msg)
	file.Close()

	mtx.Unlock()

	//再查看当前文件大小，这一步就不上锁了
	info, _ := os.Stat(aofFilePath)

	if info.Size() > AOFSize*2 {
		//如果此时没有aof重写进程在执行
		if atomic.LoadInt64(&AOFFlag) == 0 {
			go AOFRewrite()
		}
	}

}

func AOFRewrite() {
	atomic.AddInt64(&AOFFlag, 1)

	reread := AOFReread()

	file, _ := os.Create(aofFilePath)
	file.Write(reread)
	file.Close()
	//修改AOFSize,先认为reread里命令长度是重写后的长度（简化一下）
	info, _ := os.Stat(aofFilePath)
	AOFSize = info.Size()
	atomic.SwapInt64(&AOFFlag, 0)

	//把reread过程中产生的新数据存入,防止可能开始时会出现子进程创建子进程以至于一想到aofBuf功能，就不复用AOF()了
	mtx.Lock()

	//检查aof文件是否存在
	file, err := os.OpenFile(aofFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666) //权限没想好怎么设置
	if err != nil {
		log.Println("aof文件创建失败")
		//没想好这里怎么处理
		return
	}
	//写入
	file.Write(AOFBuf[:len(AOFBuf)])
	file.Close()

	mtx.Unlock()
	//大小就不查了

	//读完把aofBuf清零（重新再开一片内存）
	AOFBuf = make([]byte, 0)
}
