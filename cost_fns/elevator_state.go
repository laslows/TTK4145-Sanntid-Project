package costfns

import (
	elevator_state "./elevator_state"
	config "./config"
	"fmt"
)

type CallType int

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
	Down Dirn = iota - 1
	Stop
	Up
)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	Moving
	DoorOpen
)

type LocalElevatorState struct {
	ElevatorBehaviour Behaviour
	int               floor
	Dirn              Direction
	bool[]              cabRequests[N_FLOORS] //TODO: FIX
}

type ElevatorState struct {
	ElevatorBehaviour Behaviour
	int               floor
	Dirn              Direction
	bool[3][]           requests[N_FLOORS][N_BUTTONS] //TODO: FIX
}

func LocalElevatorState (e ElevatorState) {
	return LocalElevatorState{
		ElevatorBehaviour: e.Behaviour,
		Floor:            e.floor,
		Dirn:             e.Direction
	}
}

func ElevatorState WithRequests(e LocalElevatorState, bool[2][] hallReqs){
	return ElevatorState{
		ElevatorBehaviour: e.Behaviour,
		Floor:            e.floor,
		Dirn:             e.Direction,
		zip(hallReqs, e.cabRequests) //TODO: FIX
	}
}

func HallRequests(e elevator_state.ElevatorState) [][]bool {
    hallRequests := make([][]bool, len(e.requests))
    for floor := 0; floor < len(e.requests); floor++ {
        hallRequests[floor] = e.requests[floor][0:2]
    }
    return hallRequests