package costfns

import (
	elevator_state "./elevator_state"
	elevator_algorithm "./elevator_algorithm"
)

const (
	DoorOpenDuration  int  = 3000;
	TravelDuration  int    = 2500;
	IncludeCab bool         = false;
)

type ClearRequestType int

const (
	All ClearRequestType = iota
	InDirn
)

// TODO: Fix
ClearRequestType clearRequestType = ClearRequestType.InDirn;