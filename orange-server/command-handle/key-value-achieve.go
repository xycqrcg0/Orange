package command_handle

import (
	"orange-server/models"
	protocalutils "orange-server/utils"
)

func (database *Base) Set(key string, value string) (msg []byte, o bool) {
	keysds := models.NewSDS([]byte(key))
	valuesds := models.NewSDS([]byte(value))
	node := database.Find([]byte(key))
	if node == nil {
		database.Sum++
	}
	database.PushIn(*keysds, valuesds)
	msg = protocalutils.GenerateMsg("ok,1 key-value has been stored")
	return msg, true
}

func (database *Base) Get(key string) (msg []byte) {
	node := database.Find([]byte(key))
	if node == nil {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg
	}
	//如果这里查到的value不是SDS呢,,,要不要返回,,,
	//>选择不返回<
	value, ok := node.Value.(*models.SDS)
	if !ok {
		msg := protocalutils.GenerateMsg("the key has been used by other type")
		return msg
	}
	msg = protocalutils.GenerateMsg(string(value.Buf))
	return msg
}

func (database *Base) Delete(key string) (msg []byte, o bool) {
	//同样问题，如果不是key-value里的key，就不删了，返回未找到
	node := database.Find([]byte(key))
	if node == nil {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}
	_, ok := node.Value.(*models.SDS)
	if !ok {
		msg = protocalutils.GenerateMsg("the key has been used by other type")
		return msg, false
	}

	//key不存在才会返回false，这里就不检查了
	database.DeleteD([]byte(key))

	database.Sum--

	msg = protocalutils.GenerateMsg("ok, 1 key has been deleted")
	return msg, true
}
