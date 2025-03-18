package command_handle

import (
	"net"
	"regexp"
	"strconv"
)

//要不要对key和value做一下规范呢？先放放
//有点粗暴的分配方法~

//嘶~为什么不一开始用正则,丢给ai写表达式······

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
	}
	//zaddPattern := regexp.MustCompile()
	//zremPattern := regexp.MustCompile()
	//zrangePattern := regexp.MustCompile()
	for _, command := range commands {
		switch true {
		case patterns["set"].MatchString(command):
			params := patterns["set"].FindStringSubmatch(command)
			//params里第一个匹配到的是函数名
			Set(conn, params[1], params[2])
			return
		case patterns["get"].MatchString(command):
			params := patterns["get"].FindStringSubmatch(command)
			Get(conn, params[1])
			return
		case patterns["delete"].MatchString(command):
			params := patterns["delete"].FindStringSubmatch(command)
			Delete(conn, params[1])
			return
		case patterns["addr"].MatchString(command):
			params := patterns["addr"].FindStringSubmatch(command)
			Addr(conn, params[1], params[2])
			return
		case patterns["addl"].MatchString(command):
			params := patterns["addl"].FindStringSubmatch(command)
			Addl(conn, params[1], params[2])
			return
		case patterns["popr"].MatchString(command):
			params := patterns["popr"].FindStringSubmatch(command)
			Popr(conn, params[1])
			return
		case patterns["popl"].MatchString(command):
			params := patterns["popl"].FindStringSubmatch(command)
			Popl(conn, params[1])
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
			//case patterns["hset"].MatchString(command):
			//	params := patterns["hset"].FindStringSubmatch(command)
			//	return
			//case patterns["hget"].MatchString(command):
			//	params := patterns["hget"].FindStringSubmatch(command)
			//	return
			//case patterns["sadd"].MatchString(command):
			//	params := patterns["sadd"].FindStringSubmatch(command)
			//	return
			//case patterns["smembers"].MatchString(command):
			//	params := patterns["smembers"].FindStringSubmatch(command)
			//	return
			//case patterns["srem"].MatchString(command):
			//	params := patterns["srem"].FindStringSubmatch(command)
			//	return
		}

		Invalid(conn)
		return
	}
}
