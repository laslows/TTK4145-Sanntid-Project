package costfns

import (
	elevator_state "./elevator_state"
)

//Already have this
func requestsAbove(e elevator_state.ElevatorState) bool {
	for floor := e.floor + 1; floor < len(e.requests); floor++ {
		for _, hasRequest := range e.requests[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}


//Already have this
func requestsBelow(e elevator_state.ElevatorState) bool {
	for floor := e.floor - 1; floor >= 0; floor-- {
		for _, hasRequest := range e.requests[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}

//Checks if I have any orders at any floor
func AnyRequests(e elevator_state.ElevatorState) bool {
	for floor := 0; floor < len(e.requests); floor++ {
		for _, hasRequest := range e.requests[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}

//Checks if elevator has any requests at current floor
func AnyRequestsAtFloor(e elevator_state.ElevatorState) bool {
	for _, hasRequest := range e.requests[e.floor] {
		if hasRequest {
			return true
		}
	}
	return false
}

// TODO: Fix case-variable names and import CallType
//I think we already have this?? Maybe add checks for below 0 or higher than N_FLOORS
func ShouldStop(e elevator_state.ElevatorState) bool {
	switch e.direction {
	case Up:
		return (e.requests[e.floor][HallUp] ||
			e.requests[e.floor][Cab] ||
			!e.requestsAbove ||
			e.floor == 0 ||
			e.floor >= len(e.requests)-1)
	case Down:
		return (e.requests[e.floor][HallDown] ||
			e.requests[e.floor][Cab] ||
			!e.requestsBelow ||
			e.floor == 0 ||
			e.floor >= len(e.requests)-1)
	case Stop:
		return true
	default:
		panic("Invalid direction")
	}
}

func ChooseDirection(e elevator_state.ElevatorState) Dirn {
	switch e.direction {
	case Up:
		if requestsAbove(e) {
			return Up
		} else if AnyRequestsAtFloor(e) {
			return Stop
		} else if requestsBelow(e) {
			return Down
		} else {
			return Stop
		}
	case Down, Stop:
		if requestsBelow(e) {
			return Down
		} else if AnyRequestsAtFloor(e) {
			return Stop
		} else if requestsAbove(e) {
			return Up
		} else {
			return Stop
		}
	default:
		panic("Invalid direction")
	}
}

func ClearReqsAtFloor(
	e *elevator_state.ElevatorState,
	clearMode ClearRequestType,
	onClearedRequest func(CallType),
) {
	clearReq := func(c CallType) {
		if e.Requests[e.Floor][int(c)] {
			if onClearedRequest != nil {
				onClearedRequest(c)
			}
			e.Requests[e.Floor][int(c)] = false
		}
	}

	switch clearMode {
	case All:
		for c := 0; c < len(e.Requests[0]); c++ {
			clearReq(CallType(c))
		}

	case InDirn:
		// Always clear cab request at current floor
		clearReq(Cab)

		switch e.Direction {
		case Up:
			if e.Requests[e.Floor][int(HallUp)] {
				clearReq(HallUp)
			} else if !requestsAbove(*e) {
				clearReq(HallDown)
			}

		case Down:
			if e.Requests[e.Floor][int(HallDown)] {
				clearReq(HallDown)
			} else if !requestsBelow(*e) {
				clearReq(HallUp)
			}

		case Stop:
			clearReq(HallUp)
			clearReq(HallDown)

		default:
			panic("invalid direction")
		}

	default:
		panic("invalid clear request mode")
	}
}
