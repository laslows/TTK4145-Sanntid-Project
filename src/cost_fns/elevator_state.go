package costfns

import (
	config "./config"
)

type CallType int

HAllup

const (
	HallUp CallType = iota
	HallDown
	Cab
)

type HallCallType int

const (
	TypeUp HallCallType = iota
	TypeDown
)

type Dirn int

const (
	Down Dirn = -1
	Stop Dirn = 0
	Up   Dirn = 1
)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	Moving
	DoorOpen
)

type LocalElevatorState struct {
	Behaviour   ElevatorBehaviour
	Floor       int
	Direction   Dirn
	CabRequests [config.N_FLOORS]bool
}

type ElevatorState struct {
	Behaviour ElevatorBehaviour
	Floor     int
	Direction Dirn

	Requests [config.N_FLOORS][config.N_BUTTONS]bool
}

func ToLocalElevatorState(e ElevatorState) LocalElevatorState {
	var cab [config.N_FLOORS]bool
	for f := 0; f < config.N_FLOORS; f++ {
		cab[f] = e.Requests[f][int(Cab)]
	}

	return LocalElevatorState{
		Behaviour:   e.Behaviour,
		Floor:       e.Floor,
		Direction:   e.Direction,
		CabRequests: cab,
	}
}

func WithRequests(e LocalElevatorState, hallReqs [config.N_FLOORS][2]bool) ElevatorState {
	var reqs [config.N_FLOORS][config.N_BUTTONS]bool

	for f := 0; f < config.N_FLOORS; f++ {
		reqs[f][int(HallUp)] = hallReqs[f][0]
		reqs[f][int(HallDown)] = hallReqs[f][1]
		reqs[f][int(Cab)] = e.CabRequests[f]
	}

	return ElevatorState{
		Behaviour: e.Behaviour,
		Floor:     e.Floor,
		Direction: e.Direction,
		Requests:  reqs,
	}
}

func HallRequests(e ElevatorState) [][]bool {
	hallRequests := make([][]bool, len(e.Requests))
	for floor := 0; floor < len(e.Requests); floor++ {
		hallRequests[floor] = []bool{
			e.Requests[floor][int(HallUp)],
			e.Requests[floor][int(HallDown)],
		}
	}
	return hallRequests
}
