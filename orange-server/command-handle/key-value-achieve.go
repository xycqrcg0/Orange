package command_handle

import (
	"net"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

func Invalid(conn net.Conn) {
	msg := protocalutils.GenerateMsg("Illegal Input")
	conn.Write(msg)
}

func Set(conn net.Conn, key string, value string) bool {
	keysds := models.NewSDS([]byte(key))
	valuesds := models.NewSDS([]byte(value))
	if !data.Database.PushIn(*keysds, valuesds) {
		msg := protocalutils.GenerateMsg("key has existed")
		conn.Write(msg)
		return false
	}
	msg := protocalutils.GenerateMsg("ok,1 key-value has been stored")
	conn.Write(msg)
	return true
}

func Get(conn net.Conn, key string) {
	node := data.Database.Find([]byte(key))
	if node == nil {
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
	//如果这里查到的value不是SDS呢,,,要不要返回,,,
	//>选择不返回<
	value, ok := node.Value.(*models.SDS)
	if !ok {
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
	msg := protocalutils.GenerateMsg(string(value.Buf))
	conn.Write(msg)
	return
}

func Delete(conn net.Conn, key string) bool {
	//同样问题，如果不是key-value里的key，就不删了，返回未找到
	node := data.Database.Find([]byte(key))
	if node == nil {
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return false
	}
	_, ok := node.Value.(*models.SDS)
	if !ok {
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return false
	}

	//key不存在才会返回false，这里就不检查了
	data.Database.Delete([]byte(key))

	msg := protocalutils.GenerateMsg("ok, 1 key has been deleted")
	conn.Write(msg)
	return true
}
