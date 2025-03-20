package command_handle

import (
	"net"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

// OListNode 芝士value(list)
type OListNode struct {
	Content *models.SDS
	Left    *OListNode
	Right   *OListNode
}

//OList的存储，放进database的value是最左边的值

// Addr 右侧插入
func Addr(conn net.Conn, key string, value string) {
	newListNode := &OListNode{Left: nil, Right: nil}
	newListNode.Content = models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，只要value是list类型，就插入，不然报错
		valueList, ok := node.Value.(*OListNode)
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
	newListNode := &OListNode{Left: nil, Right: nil}
	newListNode.Content = models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，只要value是list类型，就插入，不然报错
		valueList, ok := node.Value.(*OListNode)
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

		//修改一下database里存的值
		node.Value = newListNode

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
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key is used by other type")
			conn.Write(msg)
			return
		}

		//在database里的就是最左边的值

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

func Popr(conn net.Conn, key string) {
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，再确定value是list类型
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}
		//右移
		q := valueList.Right
		if q == nil {
			//也就是说当前这个就是最右边的量（害，麻烦，还要注意修改database里存储的值）
			p := valueList.Left
			if p == nil {
				//嘶，那么这个值就是列表里唯一的值，删了
				data.Database.Delete([]byte(key))
			} else {
				p.Right = nil
				node.Value = p
			}
			msg := protocalutils.GenerateMsg("ok, pop 1 value")
			conn.Write(msg)
			return
		}
		for q.Right != nil {
			valueList = valueList.Right
			q = q.Right
		}
		//此时q就是最右端的value，那就把从右数第二个的right指针置为nil
		valueList.Right = nil

		msg := protocalutils.GenerateMsg("ok, pop 1 value")
		conn.Write(msg)
		return
	} else {
		//该key不存在
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
}

func Popl(conn net.Conn, key string) {
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，再确定value是list类型
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}

		//valueList是最左端的值
		q := valueList.Right
		if q == nil {
			//嘶，那么这个值就是列表里唯一的值，删了
			data.Database.Delete([]byte(key))
		} else {
			q.Left = nil
			node.Value = q
		}
		msg := protocalutils.GenerateMsg("ok, pop 1 value")
		conn.Write(msg)
		return
	} else {
		//该key不存在
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
}

// Lrange 从start开始读，到stop（stop数据不读取）
func Lrange(conn net.Conn, key string, start int, stop int) {
	if start >= stop || start < 0 {
		msg := protocalutils.GenerateMsg("invalid index")
		conn.Write(msg)
		return
	}

	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，再确定value是list类型
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}
		//移到最左边
		for valueList.Left != nil {
			valueList = valueList.Left
		}

		searchRange := stop - start

		//再右移到start处
		for start > 0 {
			if valueList == nil {
				break
			}
			valueList = valueList.Right
			start--
		}
		if start != 0 {
			msg := protocalutils.GenerateMsg("invalid index")
			conn.Write(msg)
			return
		}

		values := make([]string, 0)

		for searchRange > 0 {
			values = append(values, string(valueList.Content.Buf[:valueList.Content.Length]))
			valueList = valueList.Right
			searchRange--
		}
		if searchRange != 0 {
			msg := protocalutils.GenerateMsg("invalid index")
			conn.Write(msg)
			return
		}

		msg := protocalutils.GenerateMsg(values...)
		conn.Write(msg)
		return

	} else {
		//该key不存在
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
}
