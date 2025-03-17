package command_handle

import (
	"net"
	"strings"
)

//要不要对key和value做一下规范呢？先放放

func CommandsAssign(conn net.Conn, commands []string) {
	for _, command := range commands {
		l := len(command)

		//set
		if strings.HasPrefix(command, "set") {
			if command[3] != '(' || command[l-1] != ')' {
				Invalid(conn)
			}
			rawParams := command[4 : l-1]
			params := strings.Split(rawParams, ",")
			if len(params) != 2 {
				Invalid(conn)
			}

			Set(conn, params[0], params[1])
			return
		}

		//get
		if strings.HasPrefix(command, "get") {
			if command[3] != '(' || command[l-1] != ')' {
				Invalid(conn)
			}
			param := command[4 : l-1]

			Get(conn, param)
			return
		}

		//delete
		if strings.HasPrefix(command, "delete") {
			if command[6] != '(' || command[len(command)-1] != ')' {
				Invalid(conn)
			}
			param := command[4 : l-1]

			Delete(conn, param)
			return
		}

		Invalid(conn)
	}
}
