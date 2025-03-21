package command_handle

import (
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
func (database *Base) Addr(key string, value string) (msg []byte, o bool) {
	newListNode := &OListNode{Left: nil, Right: nil}
	newListNode.Content = models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，只要value是list类型，就插入
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			//覆盖
			keysds := models.NewSDS([]byte(key))
			database.PushIn(*keysds, newListNode)
		} else {
			//右侧插入
			for valueList.Right != nil {
				valueList = valueList.Right
			}
			//找到最右端了
			newListNode.Left = valueList
			valueList.Right = newListNode
		}
	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		database.PushIn(*keysds, newListNode)
	}
	msg = protocalutils.GenerateMsg("ok, 1 value has been inserted")
	return msg, true
}

// Addl 左侧插入
func (database *Base) Addl(key string, value string) (msg []byte, o bool) {
	newListNode := &OListNode{Left: nil, Right: nil}
	newListNode.Content = models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，只要value是list类型，就插入
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			//覆盖
			keysds := models.NewSDS([]byte(key))
			database.PushIn(*keysds, newListNode)
		} else {
			//左侧插入
			for valueList.Left != nil {
				valueList = valueList.Left
			}
			//找到最左端了
			newListNode.Right = valueList
			valueList.Left = newListNode

			//修改一下database里存的值
			node.Value = newListNode
		}
	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//嘶，又把key已存在的情况排除了，pushIn返回的bool又没用了···
		database.PushIn(*keysds, newListNode)

	}
	msg = protocalutils.GenerateMsg("ok, 1 value has been inserted")
	return msg, true
}

// Lindex 索引查询
func (database *Base) Lindex(key string, index int) (msg []byte) {
	//诶诶诶，就简单粗暴一点了
	node := database.Find([]byte(key))
	if node != nil {
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg = protocalutils.GenerateMsg("the key is used by other type")
			return msg
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
			msg = protocalutils.GenerateMsg("invalid index")
			return msg
		}

		msg = protocalutils.GenerateMsg(string(valueList.Content.Buf))
		return msg
	}
	msg = protocalutils.GenerateMsg("key is not existed")
	return msg
}

func (database *Base) Popr(key string) (msg []byte, o bool) {
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，再确定value是list类型
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg, false
		}
		//右移
		q := valueList.Right
		if q == nil {
			//也就是说当前这个就是最右边的量（害，麻烦，还要注意修改database里存储的值）
			p := valueList.Left
			if p == nil {
				//嘶，那么这个值就是列表里唯一的值，删了
				database.DeleteD([]byte(key))
			} else {
				p.Right = nil
				node.Value = p
			}
			msg = protocalutils.GenerateMsg("ok, pop 1 value")
			return msg, true
		}
		for q.Right != nil {
			valueList = valueList.Right
			q = q.Right
		}
		//此时q就是最右端的value，那就把从右数第二个的right指针置为nil
		valueList.Right = nil

		msg = protocalutils.GenerateMsg("ok, pop 1 value")
		return msg, true
	} else {
		//该key不存在
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}
}

func (database *Base) Popl(key string) (msg []byte, o bool) {
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，再确定value是list类型
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			return msg, false
		}

		//valueList是最左端的值
		q := valueList.Right
		if q == nil {
			//嘶，那么这个值就是列表里唯一的值，删了
			database.DeleteD([]byte(key))
		} else {
			q.Left = nil
			node.Value = q
		}
		msg := protocalutils.GenerateMsg("ok, pop 1 value")
		return msg, true
	} else {
		//该key不存在
		msg := protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}
}

// Lrange 从start开始读，到stop（stop数据不读取）
func (database *Base) Lrange(key string, start int, stop int) (msg []byte) {
	if start >= stop || start < 0 {
		msg = protocalutils.GenerateMsg("invalid index")
		return msg
	}

	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，再确定value是list类型
		valueList, ok := node.Value.(*OListNode)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg
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
			msg = protocalutils.GenerateMsg("invalid index")
			return msg
		}

		values := make([]string, 0)

		for searchRange > 0 {
			values = append(values, string(valueList.Content.Buf[:valueList.Content.Length]))
			valueList = valueList.Right
			searchRange--
		}
		if searchRange != 0 {
			msg = protocalutils.GenerateMsg("invalid index")
			return msg
		}

		msg = protocalutils.GenerateMsg(values...)
		return msg

	} else {
		//该key不存在
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg
	}
}
