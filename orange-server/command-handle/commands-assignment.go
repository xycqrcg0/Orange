package command_handle

import (
	"net"
	"orange-server/global"
	"orange-server/utils"
	"regexp"
	"strconv"
	"sync/atomic"
)

//嘶~为什么不一开始用正则,丢给ai写表达式······

//分配任务时对Record进行原子操作（毕竟还要并发）,但要注意分辨此任务有没有正确执行

var transaction = 1

var ReadPatterns = map[string]*regexp.Regexp{
	"get":      regexp.MustCompile(`^get\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"lindex":   regexp.MustCompile(`^lindex\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*(\d+)\s*\)$`),
	"lrange":   regexp.MustCompile(`^lrange\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)$`),
	"hget":     regexp.MustCompile(`^hget\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"smembers": regexp.MustCompile(`^smembers\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
}

var WritePatterns = map[string]*regexp.Regexp{
	"set":    regexp.MustCompile(`^set\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"delete": regexp.MustCompile(`^delete\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"addr":   regexp.MustCompile(`^addr\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"addl":   regexp.MustCompile(`^addl\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"popr":   regexp.MustCompile(`^popr\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"popl":   regexp.MustCompile(`^popl\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)$`),
	"hset":   regexp.MustCompile(`^hset\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"sadd":   regexp.MustCompile(`^sadd\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
	"srem":   regexp.MustCompile(`^srem\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,\s*([a-zA-Z0-9_]+)\s*\)$`),
}

var PPatterns = map[string]*regexp.Regexp{
	"aof":        regexp.MustCompile(`^(on|off)\s+AOF$`),
	"odb":        regexp.MustCompile(`^(on|off)\s+ODB$`),
	"save":       regexp.MustCompile(`^save\(\s*(\d+)\s*,\s*(\d+)\s*\)$`),
	"autocommit": regexp.MustCompile(`^set\(\s*autocommit\s*,\s*(true|false)\s*\)$`),
}

//zaddPattern := regexp.MustCompile()
//zremPattern := regexp.MustCompile()
//zrangePattern := regexp.MustCompile()

func (database *Base) ReadAssign(command string) (msg []byte) {
	msg = make([]byte, 0)
	switch true {
	case ReadPatterns["get"].MatchString(command):
		params := ReadPatterns["get"].FindStringSubmatch(command)
		msg = database.Get(params[1])

	case ReadPatterns["lindex"].MatchString(command):
		params := ReadPatterns["lindex"].FindStringSubmatch(command)
		index, _ := strconv.Atoi(params[2])
		msg = database.Lindex(params[1], index)

	case ReadPatterns["lrange"].MatchString(command):
		params := ReadPatterns["lrange"].FindStringSubmatch(command)
		start, _ := strconv.Atoi(params[2])
		stop, _ := strconv.Atoi(params[3])
		msg = database.Lrange(params[1], start, stop)

	case ReadPatterns["hget"].MatchString(command):
		params := ReadPatterns["hget"].FindStringSubmatch(command)
		msg = database.Hget(params[1], params[2])

	case ReadPatterns["smembers"].MatchString(command):
		params := ReadPatterns["smembers"].FindStringSubmatch(command)
		msg = database.Smembers(params[1])

	}
	return msg
}

func (database *Base) WriteAssign(command string) (msg []byte, ok bool) {
	ok = false
	switch true {
	case WritePatterns["set"].MatchString(command):
		params := WritePatterns["set"].FindStringSubmatch(command)
		//params里第一个匹配到的是函数名
		msg, ok = database.Set(params[1], params[2])

	case WritePatterns["delete"].MatchString(command):
		params := WritePatterns["delete"].FindStringSubmatch(command)
		msg, ok = database.Delete(params[1])

	case WritePatterns["addr"].MatchString(command):
		params := WritePatterns["addr"].FindStringSubmatch(command)
		msg, ok = database.Addr(params[1], params[2])

	case WritePatterns["addl"].MatchString(command):
		params := WritePatterns["addl"].FindStringSubmatch(command)
		msg, ok = database.Addl(params[1], params[2])

	case WritePatterns["popr"].MatchString(command):
		params := WritePatterns["popr"].FindStringSubmatch(command)
		msg, ok = database.Popr(params[1])

	case WritePatterns["popl"].MatchString(command):
		params := WritePatterns["popl"].FindStringSubmatch(command)
		msg, ok = database.Popl(params[1])

	case WritePatterns["hset"].MatchString(command):
		params := WritePatterns["hset"].FindStringSubmatch(command)
		msg, ok = database.Hset(params[1], params[2], params[3])

	case WritePatterns["sadd"].MatchString(command):
		params := WritePatterns["sadd"].FindStringSubmatch(command)
		msg, ok = database.Sadd(params[1], params[2])

	case WritePatterns["srem"].MatchString(command):
		params := WritePatterns["srem"].FindStringSubmatch(command)
		msg, ok = database.Srem(params[1], params[2])

	}
	return msg, ok
}

func (database *Base) CommandsAssign(conn net.Conn, commands []string) {
	//遍历获取的命令
	for _, command := range commands {
		//默认值
		msg := utils.GenerateMsg("Illegal Input")

		var ok = false

		switch true {
		case command == "begin":
			Transaction(conn)
			continue
		case command == "SAVE":
			SAVE(conn)
			continue
		case command == "RGSAVE":
			RGSAVE(conn)
			continue
		case PPatterns["aof"].MatchString(command):
			params := PPatterns["aof"].FindStringSubmatch(command)
			if params[1] == "on" {
				//如果之前是关的，那么这时候要重写一次aof
				if atomic.LoadInt64(&global.AOFStatus) == 0 {
					atomic.SwapInt64(&global.AOFStatus, 1)
					AOFRewrite()
				}
			} else {
				atomic.SwapInt64(&global.AOFStatus, 0)
			}
			msg = utils.GenerateMsg("ok,aof status has changed")

		case PPatterns["odb"].MatchString(command):
			params := PPatterns["odb"].FindStringSubmatch(command)
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
		case PPatterns["save"].MatchString(command):
			params := PPatterns["save"].FindStringSubmatch(command)
			a, _ := strconv.Atoi(params[1])
			b, _ := strconv.Atoi(params[2])
			Stop <- true
			go Save(a, b)
			msg = utils.GenerateMsg("ok,save rule is changed")
		case PPatterns["autocommit"].MatchString(command):
			params := PPatterns["autocommit"].FindStringSubmatch(command)
			b := params[1]
			if b == "true" {
				global.Auto = true
				msg = utils.GenerateMsg("ok,autocommit is open")
			} else {
				global.Auto = false
				msg = utils.GenerateMsg("ok,autocommit is closed")
			}
		}

		//注，开启了自动提交才能写
		if global.Auto {
			database.Mtx.Lock()
			rmsg := database.ReadAssign(command)
			wmsg, wok := database.WriteAssign(command)
			database.Mtx.Unlock()

			ok = wok
			if len(rmsg) != 0 {
				msg = rmsg
			}
			if len(wmsg) != 0 {
				msg = wmsg
			}
		} else {
			database.Mtx.Lock()
			rmsg := database.ReadAssign(command)
			if len(rmsg) != 0 {
				msg = rmsg
			}
			database.Mtx.Unlock()
			for _, pattern := range WritePatterns {
				if pattern.MatchString(command) {
					msg = utils.GenerateMsg("Please begin a transaction first")
					break
				}
			}
		}

		//如果有写入操作且成功
		if ok {
			atomic.AddInt64(&Record, 1)
			//先确定aof功能有开启
			if atomic.LoadInt64(&global.AOFStatus) != 0 {
				mmsg := utils.GenerateMsg(command)
				//重写在进行->写入缓冲区；重写未进行->写入文件
				if atomic.LoadInt64(&AOFFlag) != 0 {
					WriteInAOFBuf(mmsg)
				} else {
					AOF(mmsg[:])
				}
			}
		}

		conn.Write(msg)
	}
}
