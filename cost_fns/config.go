package costfns

const (
	DoorOpenDuration int  = 3000
	TravelDuration   int  = 2500
	IncludeCab       bool = false
)

type ClearRequestType int

const (
	All ClearRequestType = iota
	InDirn
)

var clearRequestType ClearRequestType = InDirn
