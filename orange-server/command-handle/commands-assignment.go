package command_handle

import (
	"net"
	"orange-server/utils"
	"regexp"
	"strconv"
	"sync/atomic"
)

//嘶~为什么不一开始用正则,丢给ai写表达式······

//分配任务时对Record进行原子操作（毕竟还要并发）,但要注意分辨此任务有没有正确执行

var patterns = map[string]*regexp.Regexp{
	"set":      regexp.MustCompile(`^set\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"get":      regexp.MustCompile(`^get\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"delete":   regexp.MustCompile(`^delete\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"addr":     regexp.MustCompile(`^addr\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"addl":     regexp.MustCompile(`^addl\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"popr":     regexp.MustCompile(`^popr\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"popl":     regexp.MustCompile(`^popl\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"lindex":   regexp.MustCompile(`^lindex\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*(\d+)\s*\)$`),
	"lrange":   regexp.MustCompile(`^lrange\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)$`),
	"hset":     regexp.MustCompile(`^hset\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"hget":     regexp.MustCompile(`^hget\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"sadd":     regexp.MustCompile(`^sadd\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"smembers": regexp.MustCompile(`^smembers\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"srem":     regexp.MustCompile(`^srem\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),

	"aof": regexp.MustCompile(`^(on|off)\s+AOF$`),
	"odb": regexp.MustCompile(`^(on|off)\s+ODB$`),

	"SAVE":   regexp.MustCompile(`SAVE`),
	"RGSAVE": regexp.MustCompile(`RGSAVE`),
	"save":   regexp.MustCompile(`^save\(\s*(\d+)\s*,\s*(\d+)\s*\)$`),
}

func CommandsAssign(conn net.Conn, commands []string) {
	//zaddPattern := regexp.MustCompile()
	//zremPattern := regexp.MustCompile()
	//zrangePattern := regexp.MustCompile()
	for _, command := range commands {
		if command == "begin" {
			Transaction(conn)
			return
		}

		msg := make([]byte, 0)
		var ok bool
		switch true {
		case patterns["set"].MatchString(command):
			params := patterns["set"].FindStringSubmatch(command)
			//params里第一个匹配到的是函数名
			msg, ok = Set(params[1], params[2])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["get"].MatchString(command):
			params := patterns["get"].FindStringSubmatch(command)
			msg = Get(params[1])

		case patterns["delete"].MatchString(command):
			params := patterns["delete"].FindStringSubmatch(command)
			msg, ok = Delete(conn, params[1])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["addr"].MatchString(command):
			params := patterns["addr"].FindStringSubmatch(command)
			msg, ok = Addr(params[1], params[2])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["addl"].MatchString(command):
			params := patterns["addl"].FindStringSubmatch(command)
			msg, ok = Addl(params[1], params[2])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["popr"].MatchString(command):
			params := patterns["popr"].FindStringSubmatch(command)
			msg, ok = Popr(params[1])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["popl"].MatchString(command):
			params := patterns["popl"].FindStringSubmatch(command)
			msg, ok = Popl(params[1])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["lindex"].MatchString(command):
			params := patterns["lindex"].FindStringSubmatch(command)
			index, _ := strconv.Atoi(params[2])
			msg = Lindex(params[1], index)

		case patterns["lrange"].MatchString(command):
			params := patterns["lrange"].FindStringSubmatch(command)
			start, _ := strconv.Atoi(params[2])
			stop, _ := strconv.Atoi(params[3])
			msg = Lrange(params[1], start, stop)

		case patterns["hset"].MatchString(command):
			params := patterns["hset"].FindStringSubmatch(command)
			msg, ok = Hset(params[1], params[2], params[3])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["hget"].MatchString(command):
			params := patterns["hget"].FindStringSubmatch(command)
			msg = Hget(params[1], params[2])

		case patterns["sadd"].MatchString(command):
			params := patterns["sadd"].FindStringSubmatch(command)
			msg, ok = Sadd(params[1], params[2])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["smembers"].MatchString(command):
			params := patterns["smembers"].FindStringSubmatch(command)
			msg = Smembers(params[1])

		case patterns["srem"].MatchString(command):
			params := patterns["srem"].FindStringSubmatch(command)
			msg, ok = Srem(params[1], params[2])
			if ok {
				atomic.AddInt64(&Record, 1)
				//先确定aof功能有开启
				if atomic.LoadInt64(&AOFStatus) != 0 {
					mmsg := utils.GenerateMsg(command)
					//重写在进行->写入缓冲区；重写未进行->写入文件
					if atomic.LoadInt64(&AOFFlag) != 0 {
						WriteInAOFBuf(mmsg)
					} else {
						AOF(mmsg[:])
					}
				}
			}

		case patterns["aof"].MatchString(command):
			params := patterns["aof"].FindStringSubmatch(command)
			if params[1] == "on" {
				//如果之前是关的，那么这时候要重写一次aof
				if atomic.LoadInt64(&AOFStatus) == 0 {
					atomic.SwapInt64(&AOFStatus, 1)
					AOFRewrite()
				}
			} else {
				atomic.SwapInt64(&AOFStatus, 0)
			}
			msg = utils.GenerateMsg("ok,aof status has changed")
		case patterns["odb"].MatchString(command):
			params := patterns["odb"].FindStringSubmatch(command)
			if params[1] == "on" {
				atomic.SwapInt64(&ODBStatus, 1)
			} else {
				//如果之前是开的关上自动触发的save
				if atomic.LoadInt64(&ODBStatus) != 0 {
					Stop <- true
					atomic.SwapInt64(&ODBStatus, 0)
				}
			}
			msg = utils.GenerateMsg("ok,odb status has changed")

		case patterns["SAVE"].MatchString(command):
			SAVE(conn)
			return
		case patterns["RGSAVE"].MatchString(command):
			RGSAVE(conn)
			return
		case patterns["save"].MatchString(command):
			params := patterns["save"].FindStringSubmatch(command)
			a, _ := strconv.Atoi(params[1])
			b, _ := strconv.Atoi(params[2])
			Stop <- true
			go Save(a, b)
			msg = utils.GenerateMsg("ok,save rule is changed")
		}
		if len(msg) == 0 {
			msg = utils.GenerateMsg("Illegal Input")
		}
		conn.Write(msg)
	}
}
