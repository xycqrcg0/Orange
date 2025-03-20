package command_handle

import (
	"hash/fnv"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

//这个和hash差不多，稍微修改一下就行（代码就不复用了）

// OSet (芝士value(set))
type OSet struct {
	Length int
	Sum    int
	Value  []*models.SDS
}

func Sadd(key string, value string) (msg []byte, o bool) {
	valuesds := models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg, false
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % valueOSet.Length

		if valueOSet.Value[hashed] != nil {
			//因为set要求字符串是唯一的，那么当前这种情况是不被允许的（hash可以很快发现这个问题，这也是用哈希表实现的原因）
			msg = protocalutils.GenerateMsg("value has been existed")
			return msg, false
		} else {
			//这就可以放了
			valueOSet.Value[hashed] = valuesds
			//OSet里的sum 要++
			valueOSet.Sum++
		}

	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//切片先开多大呢？（这是一个问题）先小一点吧
		newValueOSet := &OSet{
			Length: 128,
			Sum:    0,
			Value:  make([]*models.SDS, 128),
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % newValueOSet.Length
		newValueOSet.Value[hashed] = valuesds
		newValueOSet.Sum++

		//把newValueOHash往database里放
		data.Database.PushIn(*keysds, newValueOSet)
	}

	msg = protocalutils.GenerateMsg("ok, 1 value has been inserted")
	return msg, true
}

func Smembers(key string) (msg []byte) {
	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
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
				values = append(values, string(value.Buf[:value.Length]))
			}
		}
		msg = protocalutils.GenerateMsg(values...)
		return msg
	} else {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg
	}
}

func Srem(key string, value string) (msg []byte, o bool) {
	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
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
		return msg, true

	} else {
		msg := protocalutils.GenerateMsg("key is not existed")
		return msg, false
	}
}
