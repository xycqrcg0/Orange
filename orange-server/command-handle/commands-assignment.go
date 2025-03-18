package command_handle

import (
	"net"
	"strconv"
	"strings"
)

//要不要对key和value做一下规范呢？先放放
//有点粗暴的分配方法~

func CommandsAssign(conn net.Conn, commands []string) {
	for _, command := range commands {
		l := len(command)

		//set
		if strings.HasPrefix(command, "set") {
			if l < 8 || command[3] != '(' || command[l-1] != ')' {
				Invalid(conn)
				return
			}
			rawParams := command[4 : l-1]
			params := strings.Split(rawParams, ",")
			if len(params) != 2 {
				Invalid(conn)
				return
			}

			Set(conn, params[0], params[1])
			return
		}

		//get
		if strings.HasPrefix(command, "get") {
			if l < 6 || command[3] != '(' || command[l-1] != ')' {
				Invalid(conn)
				return
			}
			param := command[4 : l-1]

			Get(conn, param)
			return
		}

		//delete
		if strings.HasPrefix(command, "delete") {
			if l < 9 || command[6] != '(' || command[len(command)-1] != ')' {
				Invalid(conn)
				return
			}
			param := command[7 : l-1]

			Delete(conn, param)
			return
		}

		//addr
		if strings.HasPrefix(command, "addr") {
			if l < 9 || command[4] != '(' || command[l-1] != ')' {
				Invalid(conn)
				return
			}
			rawParams := command[5 : l-1]
			params := strings.Split(rawParams, ",")
			if len(params) != 2 {
				Invalid(conn)
				return
			}

			Addr(conn, params[0], params[1])
			return
		}

		//addl
		if strings.HasPrefix(command, "addl") {
			if l < 9 || command[4] != '(' || command[l-1] != ')' {
				Invalid(conn)
				return
			}
			rawParams := command[5 : l-1]
			params := strings.Split(rawParams, ",")
			if len(params) != 2 {
				Invalid(conn)
				return
			}

			Addl(conn, params[0], params[1])
			return
		}

		//lindex
		if strings.HasPrefix(command, "lindex") {
			if l < 11 || command[6] != '(' || command[l-1] != ')' {
				Invalid(conn)
				return
			}
			rawParams := command[7 : l-1]
			params := strings.Split(rawParams, ",")
			if len(params) != 2 {
				Invalid(conn)
				return
			}

			index, err := strconv.Atoi(params[1])
			if err != nil {
				Invalid(conn)
				return
			}
			Lindex(conn, params[0], index)
			return
		}

		Invalid(conn)
	}
}
