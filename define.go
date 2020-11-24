package main

type CbId struct {
	Idx uint32
	Val uint32
}

type CnMsg struct {
	Id   CbId
	Seq  uint32
	Ack  uint32
	Len  uint16
	Flag uint16
}
