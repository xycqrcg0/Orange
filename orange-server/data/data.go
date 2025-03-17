package data

import "orange-server/models"

//存储方式

type OString struct {
	Key   models.SDS
	Value models.SDS
}
type OList struct {
	Key   models.SDS
	Value interface{}
}
type OHash struct {
	Key   models.SDS
	Value interface{}
}
type OSet struct {
	Key   models.SDS
	Value interface{}
}
type OZSet struct {
	Key   models.SDS
	Value interface{}
}

//所有数据放进对应的切片里
//在main里先初始化

var OStrings []OString
var OLists []OList
var OHashes []OHash
var OSets []OSet
var OZSets []OZSet
