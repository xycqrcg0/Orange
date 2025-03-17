package command_handle

import (
	"net"
	protocalutils "orange-server/utils"
)

func Invalid(conn net.Conn) {
	msg := protocalutils.GenerateMsg("Illegal Input")
	conn.Write(msg)
}

func Set(conn net.Conn, key string, value string) {}

func Get(conn net.Conn, key string) {}

func Delete(conn net.Conn, key string) {}
