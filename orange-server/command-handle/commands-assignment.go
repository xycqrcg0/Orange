package command_handle

import (
	"net"
	"orange-server/utils"
	"regexp"
	"strconv"
	"sync/atomic"
)

//要不要对key和value做一下规范呢？先放放
//有点粗暴的分配方法~

//嘶~为什么不一开始用正则,丢给ai写表达式······

//分配任务时对Record进行原子操作（毕竟还要并发）,但要注意分辨此任务有没有正确执行

func CommandsAssign(conn net.Conn, commands []string) {
	patterns := map[string]*regexp.Regexp{
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

		"SAVE":   regexp.MustCompile(`SAVE`),
		"RGSAVE": regexp.MustCompile(`RGSAVE`),
		"save":   regexp.MustCompile(`^save\(\s*(\d+)\s*,\s*(\d+)\s*\)$`),
	}
	//zaddPattern := regexp.MustCompile()
	//zremPattern := regexp.MustCompile()
	//zrangePattern := regexp.MustCompile()
	for _, command := range commands {
		switch true {
		case patterns["set"].MatchString(command):
			params := patterns["set"].FindStringSubmatch(command)
			//params里第一个匹配到的是函数名
			if Set(conn, params[1], params[2]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["get"].MatchString(command):
			params := patterns["get"].FindStringSubmatch(command)
			Get(conn, params[1])
			return
		case patterns["delete"].MatchString(command):
			params := patterns["delete"].FindStringSubmatch(command)
			if Delete(conn, params[1]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["addr"].MatchString(command):
			params := patterns["addr"].FindStringSubmatch(command)
			if Addr(conn, params[1], params[2]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["addl"].MatchString(command):
			params := patterns["addl"].FindStringSubmatch(command)
			if Addl(conn, params[1], params[2]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["popr"].MatchString(command):
			params := patterns["popr"].FindStringSubmatch(command)
			if Popr(conn, params[1]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["popl"].MatchString(command):
			params := patterns["popl"].FindStringSubmatch(command)
			if Popl(conn, params[1]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["lindex"].MatchString(command):
			params := patterns["lindex"].FindStringSubmatch(command)
			index, _ := strconv.Atoi(params[2])
			Lindex(conn, params[1], index)
			return
		case patterns["lrange"].MatchString(command):
			params := patterns["lrange"].FindStringSubmatch(command)
			start, _ := strconv.Atoi(params[2])
			stop, _ := strconv.Atoi(params[3])
			Lrange(conn, params[1], start, stop)
			return
		case patterns["hset"].MatchString(command):
			params := patterns["hset"].FindStringSubmatch(command)
			if Hset(conn, params[1], params[2], params[3]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["hget"].MatchString(command):
			params := patterns["hget"].FindStringSubmatch(command)
			Hget(conn, params[1], params[2])
			return
		case patterns["sadd"].MatchString(command):
			params := patterns["sadd"].FindStringSubmatch(command)
			if Sadd(conn, params[1], params[2]) {
				atomic.AddInt64(&Record, 1)
			}
			return
		case patterns["smembers"].MatchString(command):
			params := patterns["smembers"].FindStringSubmatch(command)
			Smembers(conn, params[1])
			return
		case patterns["srem"].MatchString(command):
			params := patterns["srem"].FindStringSubmatch(command)
			if Srem(conn, params[1], params[2]) {
				atomic.AddInt64(&Record, 1)
			}
			return

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
			msg := utils.GenerateMsg("ok,save rule is changed")
			conn.Write(msg)
			return

		}
		Invalid(conn)
		return
	}
}
