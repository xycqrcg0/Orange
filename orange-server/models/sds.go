package models

//按redis实现，这里搞了好几种sds，length和alloc大小不同
//这里嫌麻烦，（其实是没想好怎么把这几种抽象为一种数据类型），就只设计一种了
//ps:整体功能里似乎没有体现出使用SDS的作用吧()

type SDS struct {
	Length int    //当前数据长度
	Alloc  int    //可存储数据长度
	Buf    []byte //数据string
}

func NewSDS(s []byte) *SDS {
	//规则：if len(s) < 1024 -> 分配2*len(s)空间
	//if len(s) > 1024 -> 分配len(s) + 1024 空间
	l := len(s)
	var length = l
	var alloc int
	if length < 1024 {
		alloc = length * 2
	} else {
		alloc = length + 1024
	}
	buf := make([]byte, alloc)
	copy(buf, s[:length])

	return &SDS{
		Length: length,
		Alloc:  alloc,
		Buf:    buf,
	}
}
