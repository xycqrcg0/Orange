package command_handle

import (
	"hash/fnv"
	"orange-server/models"
)

//存储方式
//妥协了，学redis用哈希吧先

var DB *Base

type ONode struct {
	Key   models.SDS
	Value interface{}
	Next  *ONode
}

//喵的，把Int当做4字节了···（就当它是吧）

type Base struct {
	Sum    int //当前存放数据量
	Length int //data数组长度
	Max    int //最大位(目前只是摆设)
	Data   []*ONode
}

func (database *Base) PushIn(key models.SDS, value interface{}) bool {
	//数组扩容规则：每当sum和length达到某个比值时，对data进行扩容
	//先省略

	node := &ONode{Key: key, Value: value, Next: nil}

	//对key进行哈希
	h := fnv.New32a()
	h.Write(key.Buf[:key.Length])
	//要模一下别访问非法内存了
	hashKey := int(h.Sum32()) % database.Length

	newKey := string(key.Buf)

	//解决哈希冲突
	if database.Data[hashKey] != nil {
		p := database.Data[hashKey]
		for p.Next != nil {
			if string(p.Key.Buf) == newKey {
				return false
			}
			p = p.Next
		}
		if string(p.Key.Buf) == newKey {
			//这就是两个一样的key了，是直接覆盖还是先返回报错呢
			//报错
			return false
		}
		p.Next = node
	} else {
		database.Data[hashKey] = node
	}

	//这个++在将来支持并发时要注意并发问题
	database.Sum++

	return true
}

func (database *Base) DeleteD(byteKey []byte) bool {
	//要注意！！这里的byteKey与存储的SDS里的byte长度不同，比较时要处理一下······

	//对key进行哈希
	h := fnv.New32a()
	h.Write(byteKey)
	hashKey := int(h.Sum32()) % database.Length

	key := string(byteKey)

	p := database.Data[hashKey]
	if p == nil {
		//该key不存在
		return false
	}
	if q := p.Next; q != nil {
		//data[hashKey]:[][][][][p][q][][][][]...
		for q != nil {
			if string(q.Key.Buf[:q.Key.Length]) == key {
				p.Next = q.Next
				//跳过的那个元素相当于被删除，go的GC机制应该会把它删了
				database.Sum--
				return true
			}
			p = p.Next
			q = q.Next
		}
	} else {
		//data[hashKey]:[p]
		if string(p.Key.Buf[:p.Key.Length]) == key {
			database.Data[hashKey] = nil
			database.Sum--
			//同理
			return true
		}
	}
	return false
}

func (database *Base) Find(byteKey []byte) *ONode {
	//这里感觉直接给SDS里的[]byte就够了

	//对key进行哈希
	h := fnv.New32a()
	h.Write(byteKey)
	hashKey := int(h.Sum32()) % database.Length

	key := string(byteKey)

	p := database.Data[hashKey]
	if p == nil {
		//该key不存在
		return nil
	}
	if q := p.Next; q != nil {
		//data[hashKey]:[][][][][p][q][][][][]...
		for q != nil {
			if string(q.Key.Buf[:q.Key.Length]) == key {
				return q
			}
			p = p.Next
			q = q.Next
		}
	} else {
		//data[hashKey]:[p]
		if string(p.Key.Buf[:p.Key.Length]) == key {
			return p
		}
	}
	return nil
}
