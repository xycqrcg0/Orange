package command_handle

import (
	"hash/fnv"
	"net"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

//这个和hash差不多，稍微修改一下就行（代码就不复用了）

// OSet (芝士value(set))
type OSet struct {
	length int
	sum    int
	value  []*models.SDS
}

func Sadd(conn net.Conn, key string, value string) {
	valuesds := models.NewSDS([]byte(value))

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % valueOSet.length

		if valueOSet.value[hashed] != nil {
			//因为set要求字符串是唯一的，那么当前这种情况是不被允许的（hash可以很快发现这个问题，这也是用哈希表实现的原因）
			msg := protocalutils.GenerateMsg("value has been existed")
			conn.Write(msg)
			return
		} else {
			//这就可以放了
			valueOSet.value[hashed] = valuesds
			//OSet里的sum 要++
			valueOSet.sum++
		}

	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//切片先开多大呢？（这是一个问题）先小一点吧
		newValueOSet := &OSet{
			length: 128,
			sum:    0,
			value:  make([]*models.SDS, 128),
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % newValueOSet.length
		newValueOSet.value[hashed] = valuesds
		newValueOSet.sum++

		//把newValueOHash往database里放
		data.Database.PushIn(*keysds, newValueOSet)
	}

	msg := protocalutils.GenerateMsg("ok, 1 value has been inserted")
	conn.Write(msg)
	return
}

func Smembers(conn net.Conn, key string) {
	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}
		values := make([]string, 0)
		for _, value := range valueOSet.value {
			if value != nil {
				values = append(values, string(value.Buf[:value.Length]))
			}
		}
		msg := protocalutils.GenerateMsg(values...)
		conn.Write(msg)
		return
	} else {
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
}

func Srem(conn net.Conn, key string, value string) {
	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是OSet(set)类型
		valueOSet, ok := node.Value.(*OSet)
		if !ok {
			msg := protocalutils.GenerateMsg("the key has been used by other type")
			conn.Write(msg)
			return
		}

		h := fnv.New32a()
		h.Write([]byte(value))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % valueOSet.length

		if valueOSet.value[hashed] == nil {
			msg := protocalutils.GenerateMsg("value is not existed")
			conn.Write(msg)
			return
		}

		valueOSet.value[hashed] = nil
		valueOSet.sum--

		//其实真要说这里也是要考虑一下length是不是要缩减()

		msg := protocalutils.GenerateMsg("ok, 1 value has been deleted")
		conn.Write(msg)
		return

	} else {
		msg := protocalutils.GenerateMsg("key is not existed")
		conn.Write(msg)
		return
	}
}
