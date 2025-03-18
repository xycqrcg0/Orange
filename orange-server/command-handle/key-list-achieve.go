package command_handle

import (
	"net"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

//OList的存储，放进database的value是第一个插入的值

// Addr 右侧插入
func Addr(conn net.Conn, key string, value string) {
	newListNode := &models.OListNode{Left: nil, Right: nil}
	newListNode.Content = models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，只要value是list类型，就插入，不然报错
		valueList, ok := node.Value.(*models.OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}
		//右侧插入
		for valueList.Right != nil {
			valueList = valueList.Right
		}
		//找到最右端了
		newListNode.Left = valueList
		valueList.Right = newListNode

		msg := protocalutils.GenerateMsg("ok, 1 value has been inserted")
		conn.Write(msg)
		return
	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//嘶，又把key已存在的情况排除了，pushIn返回的bool又没用了···
		data.Database.PushIn(*keysds, newListNode)

		msg := protocalutils.GenerateMsg("ok, 1 value has been inserted")
		conn.Write(msg)
		return
	}
}

// Addl 左侧插入
func Addl(conn net.Conn, key string, value string) {
	newListNode := &models.OListNode{Left: nil, Right: nil}
	newListNode.Content = models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，只要value是list类型，就插入，不然报错
		valueList, ok := node.Value.(*models.OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}
		//左侧插入
		for valueList.Left != nil {
			valueList = valueList.Left
		}
		//找到最左端了
		newListNode.Right = valueList
		valueList.Left = newListNode

		msg := protocalutils.GenerateMsg("ok, 1 value has been inserted")
		conn.Write(msg)
		return
	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//嘶，又把key已存在的情况排除了，pushIn返回的bool又没用了···
		data.Database.PushIn(*keysds, newListNode)

		msg := protocalutils.GenerateMsg("ok, 1 value has been inserted")
		conn.Write(msg)
		return
	}
}

// Lindex 索引查询
func Lindex(conn net.Conn, key string, index int) {
	//诶诶诶，就简单粗暴一点了
	node := data.Database.Find([]byte(key))
	if node != nil {
		valueList, ok := node.Value.(*models.OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key is used by other type")
			conn.Write(msg)
			return
		}
		for valueList.Left != nil {
			valueList = valueList.Left
		}

		for index > 0 {
			if valueList.Right == nil {
				break
			}
			valueList = valueList.Right
			index--
		}
		if index != 0 {
			//没有正常遍历到目标位置
			msg := protocalutils.GenerateMsg("invalid index")
			conn.Write(msg)
			return
		}

		msg := protocalutils.GenerateMsg(string(valueList.Content.Buf))
		conn.Write(msg)
		return
	}
	msg := protocalutils.GenerateMsg("key is not existed")
	conn.Write(msg)
}
