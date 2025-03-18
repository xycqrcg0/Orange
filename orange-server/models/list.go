package models

type OListNode struct {
	Content *SDS
	Left    *OListNode
	Right   *OListNode
}
