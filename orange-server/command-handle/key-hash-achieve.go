package command_handle

import (
	"hash/fnv"
	"orange-server/data"
	"orange-server/models"
	protocalutils "orange-server/utils"
)

//data那里的hash似乎不太好复用

type OHashNode struct {
	Field models.SDS
	Value *models.SDS
	Next  *OHashNode
}

// OHash (芝士value(hash))
type OHash struct {
	Length int
	Sum    int
	Value  []*OHashNode
}

func Hset(key string, field string, value string) (msg []byte, o bool) {
	fieldsds := models.NewSDS([]byte(field))
	valuesds := models.NewSDS([]byte(value))
	newHashNode := &OHashNode{Field: *fieldsds, Value: valuesds, Next: nil}

	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是hashNode切片(hash)类型
		valueOHash, ok := node.Value.(*OHash)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg, false
		}
		//准备把数据放入

		//别忘了要做扩容的准备
		//扩容操作暂放

		//找到这个newHashNode该放哪里
		h := fnv.New32a()
		h.Write([]byte(field))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % valueOHash.Length
		//解决哈希冲突
		p := valueOHash.Value[hashed]
		if p != nil {
			for p.Next != nil {
				p = p.Next
			}
			if string(p.Field.Buf[:p.Field.Length]) == field {
				//那么该field是重复了，报错
				msg = protocalutils.GenerateMsg("field has been existed")
				return msg, false
			}
			//在末尾放上
			p.Next = newHashNode
		} else {
			//直接放
			valueOHash.Value[hashed] = newHashNode
		}
		valueOHash.Sum++

	} else {
		//那就新建这个key-value
		keysds := models.NewSDS([]byte(key))
		//切片先开多大呢？（这是一个问题）先小一点吧
		newValueOHash := &OHash{
			Length: 128,
			Sum:    0,
			Value:  make([]*OHashNode, 128),
		}

		//把newHashNode放进newValueHash里,别忘了哈希时不要加上byte后面的空byte（直接用field吧）
		h := fnv.New32a()
		h.Write([]byte(field))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % newValueOHash.Length
		//毕竟是新开的切片了，哈希冲突是不存在的
		newValueOHash.Value[hashed] = newHashNode
		newValueOHash.Sum++

		//把newValueOHash往database里放
		data.Database.PushIn(*keysds, newValueOHash)
	}

	msg = protocalutils.GenerateMsg("ok, 1 field-value has been inserted")
	return msg, true
}

func Hget(key string, field string) (msg []byte) {
	//先看看该key是否存在
	node := data.Database.Find([]byte(key))
	if node != nil {
		//该key存在，确定该value是hashNode切片(hash)类型
		valueOHash, ok := node.Value.(*OHash)
		if !ok {
			msg = protocalutils.GenerateMsg("the key has been used by other type")
			return msg
		}
		h := fnv.New32a()
		h.Write([]byte(field))
		//要模一下别访问非法内存了
		hashed := int(h.Sum32()) % valueOHash.Length

		p := valueOHash.Value[hashed]
		if p != nil {
			if string(p.Field.Buf[:p.Field.Length]) == field {
				msg = protocalutils.GenerateMsg(string(p.Value.Buf[:p.Value.Length]))
				return msg
			}
			for p.Next != nil {
				if string(p.Field.Buf[:p.Field.Length]) == field {
					//找到了
					msg = protocalutils.GenerateMsg(string(p.Value.Buf[:p.Value.Length]))
					return msg
				}
				p = p.Next
			}
		}
		msg = protocalutils.GenerateMsg("field is not existed")
		return msg
	} else {
		msg = protocalutils.GenerateMsg("key is not existed")
		return msg
	}
}
