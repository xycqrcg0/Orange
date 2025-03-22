package command_handle

import (
	"hash/fnv"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

//这个和hash差不多，稍微修改一下就行（代码就不复用了）

// OSet (芝士value(set))
type OSet struct {
	Length int
	Sum    int
	Value  []*OSetNode
}

type OSetNode struct {
	Value *models.SDS
	Next  *OSetNode
}

func (database *Base) Sadd(key string, value string) (msg []byte, o bool) {
	valuesds := models.NewSDS([]byte(value))

	newSetNode := &OSetNode{Value: valuesds, Next: nil}

	//先看看该key是否存在
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			//直接覆盖
			newValueOSet := &OSet{
				Length: 128,
				Sum:    0,
				Value:  make([]*OSetNode, 128),
			}
			h := fnv.New32a()
			h.Write([]byte(value))
			//要模一下别访问非法内存了
			hashed := int(h.Sum32()) % newValueOSet.Length
			newValueOSet.Value[hashed] = newSetNode
			newValueOSet.Sum++
			node.Value = newValueOSet
		} else {
			h := fnv.New32a()
			h.Write([]byte(value))
			//要模一下别访问非法内存了
			hashed := int(h.Sum32()) % valueOSet.Length

			if valueOSet.Value[hashed] != nil {
				//因为set要求字符串是唯一的，那么当前这种情况很可能是value重复了（hash可以很快发现这个问题，这也是用哈希表实现的原因），但也可能是单纯哈到一起了，要检验
				p := valueOSet.Value[hashed]
				for p.Next != nil {
					if string(p.Value.Buf) == value {
						//那确实是value重复了，就不动了
						msg = protocalutils.GenerateMsg("value has been existed")
						return msg, false
					}
					p = p.Next
				}
				if string(p.Value.Buf) == value {
					msg = protocalutils.GenerateMsg("value has been existed")
					return msg, false
				}
				p.Next = newSetNode
				//OSet里的sum 要++
				valueOSet.Sum++
			} else {
				//这就可以放了
				valueOSet.Value[hashed] = newSetNode
				//OSet里的sum 要++
				valueOSet.Sum++
			}
		}
	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//切片先开多大呢？（这是一个问题）先小一点吧
		newValueOSet := &OSet{
			Length: 128,
			Sum:    0,
			Value:  make([]*OSetNode, 128),
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % newValueOSet.Length
		newValueOSet.Value[hashed] = newSetNode
		newValueOSet.Sum++

		//把newValueOSet往database里放
		database.PushIn(*keysds, newValueOSet)

		database.Sum++
	}

	msg = protocalutils.GenerateMsg("ok, 1 value has been inserted")
	return msg, true
}

func (database *Base) Smembers(key string) (msg []byte) {
	//先看看该key是否存在
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg
		}
		values := make([]string, 0)
		for _, value := range valueOSet.Value {
			if value != nil {
				values = append(values, string(value.Value.Buf[:value.Value.Length]))
			}
		}
		msg = protocalutils.GenerateMsg(values...)
		return msg
	} else {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg
	}
}

func (database *Base) Srem(key string, value string) (msg []byte, o bool) {
	//先看看该key是否存在
	node := database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg, false
		}

		if valueOSet.Sum == 1 {
			//那么这个元素删了就没值了，键值对是不是也要删了
			node = nil
			msg = protocalutils.GenerateMsg("ok, 1 value has been deleted")
			return msg, true
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % valueOSet.Length

		if valueOSet.Value[hashed] == nil {
			msg = protocalutils.GenerateMsg("value is not existed")
			return msg, false
		}

		valueOSet.Value[hashed] = nil
		valueOSet.Sum--

		//其实真要说这里也是要考虑一下length是不是要缩减()

		msg = protocalutils.GenerateMsg("ok, 1 value has been deleted")

		database.Sum--

		return msg, true

	} else {
		msg := protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}
}
