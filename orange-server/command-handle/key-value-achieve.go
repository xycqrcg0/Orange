package command_handle

import (
	"net"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

func Set(key string, value string) (msg []byte, o bool) {
	keysds := models.NewSDS([]byte(key))
	valuesds := models.NewSDS([]byte(value))
	if !data.Database.PushIn(*keysds, valuesds) {
		msg := protocalutils.GenerateMsg("key has existed")
		return msg, false
	}
	msg = protocalutils.GenerateMsg("ok,1 key-value has been stored")
	return msg, true
}

func Get(key string) (msg []byte) {
	node := data.Database.Find([]byte(key))
	if node == nil {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg
	}
	//如果这里查到的value不是SDS呢,,,要不要返回,,,
	//>选择不返回<
	value, ok := node.Value.(*models.SDS)
	if !ok {
		msg := protocalutils.GenerateMsg("key is not existed")
		return msg
	}
	msg = protocalutils.GenerateMsg(string(value.Buf))
	return msg
}

func Delete(conn net.Conn, key string) (msg []byte, o bool) {
	//同样问题，如果不是key-value里的key，就不删了，返回未找到
	node := data.Database.Find([]byte(key))
	if node == nil {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}
	_, ok := node.Value.(*models.SDS)
	if !ok {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}

	//key不存在才会返回false，这里就不检查了
	data.Database.Delete([]byte(key))

	msg = protocalutils.GenerateMsg("ok, 1 key has been deleted")
	return msg, true
}
