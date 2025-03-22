package command_handle

import (
	"log"
	"net"
	"orange-server/global"
	protocalutils "orange-server/utils"
	"sync/atomic"
	"time"
)

// SAVE 阻塞式保存
func SAVE(conn net.Conn) {
	if atomic.LoadInt64(&global.ODBStatus) == 0 {
		msg := protocalutils.GenerateMsg("ODB is not enabled")
		conn.Write(msg)
		return
	}

	//在此之前要看看当前有没有SAVE进程在执行（相当于锁吧）（原子读取）
	if atomic.LoadInt64(&SAVEFlag) != 0 {
		msg := protocalutils.GenerateMsg("please wait,1 (RG)SAVE routine is running")
		conn.Write(msg)
		return
	}

	//再等等自动触发的save
	for atomic.LoadInt64(&SaveF) != 0 {
	}

	//加“锁”
	atomic.AddInt64(&SAVEFlag, 1)

	r := atomic.LoadInt64(&Record)

	if err := WriteODB(); err != nil {
		msg := protocalutils.GenerateMsg("SAVE failure :" + err.Error())
		conn.Write(msg)
		return
	}

	atomic.AddInt64(&Record, -1*r)

	//解“锁”
	atomic.SwapInt64(&SAVEFlag, 0)

	msg := protocalutils.GenerateMsg("OK, 1 .odb file has been created")
	conn.Write(msg)
	return
}

// RGSAVE 非阻塞式保存
func RGSAVE(conn net.Conn) {
	if atomic.LoadInt64(&global.ODBStatus) == 0 {
		msg := protocalutils.GenerateMsg("ODB is not enabled")
		conn.Write(msg)
		return
	}

	//汗
	go SAVE(conn)
	//就害怕并发时该进程写入的msg会破坏结构
	//but write方法好像是并发安全的()
}

// Save 每次使用前先往Stop里塞个量，然后开进程使用 （初始设置时除外）
func Save(a int, b int) {
	t := time.Second * time.Duration(a)
	ticker := time.NewTicker(t)
	defer ticker.Stop()

	for {
		select {
		case <-Stop:
			//表示有新的Save规则了，这个要停了
			return
		case <-ticker.C:
			//看看有没有b个键被修改（原子读取）
			r := atomic.LoadInt64(&Record)
			if r >= int64(b) {
				//自动触发的话，如果当前有手动触发，就不进行了？要进行，等SAVE结束
				//该自动触发的写入要不要对手动触发的SAVE进行阻塞呢？不阻塞，让它等等

				//如果当前有SAVE正在进行，就停下等等
				for atomic.LoadInt64(&SAVEFlag) != 0 {
				}

				atomic.AddInt64(&SaveF, 1)

				if err := WriteODB(); err != nil {
					//修改失败
					log.Println("修改失败,err:", err)
				} else {
					//修改成功，改下Record的数据（其实这里的r要比实际写入的数据量要小一些，暂且不考虑这点，这只会导致写入比实际情况更频繁）
					atomic.AddInt64(&Record, -1*r)
				}

				atomic.SwapInt64(&SaveF, 0)

				log.Println("save success")
			}
		}
	}
}
