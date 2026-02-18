package costfns

import (
	elevator_state "./elevator_state"
)

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

func AnyRequestsAtFloor(e elevator_state.ElevatorState) bool {
	for _, hasRequest := range e.requests[e.floor] {
		if hasRequest {
			return true
		}
	}
	return false
}

// TODO: Fix case-variable names and import CallType
func ShouldStop(e elevator_state.ElevatorState) bool {
	switch e.direction {
	case Up:
		return (e.requests[e.floor][CallType.hallUp] ||
			e.requests[e.floor][CallType.cab] ||
			!e.requestsAbove ||
			e.floor == 0 ||
			e.floor >= len(e.requests)-1)
	case Down:
		return (e.requests[e.floor][CallType.hallDown] ||
			e.requests[e.floor][CallType.cab] ||
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

func ClearReqsAtFloor(e *elevator_state.ElevatorState, delegate( CallType c) OnClearedRequest = null){
	auto e2 = e;

	clear(Calltype c) void{
		if e2.requests[e2.floor][c] {
			if &OnClearedRequest {
				OnClearedRequest(c)
			}
		e2.requests[e2.floor][c] = false
		}
	}

	switch ClearRequestType {
	case All:
		for c := CallType.min; c < len(e2.requests[0]); c++ {
			clear(c)
		}
		break
	case InDirn:
		clear CallType.cab

		switch e.Direction {
		case Up:
			if e2.requests[e2.floor][CallType.hallUp] {
				clear(CallType.hallUp)
			} else if (!e2.requestsAbove) {
				clear(CallType.hallDown)
			}
			break
		case Down:
			if e2.requests[e2.floor][CallType.hallDown] {
				clear(CallType.hallDown)
			} else if (!e2.requestsBelow) {
				clear(CallType.hallUp)
			}
			break
		case Stop:
			clear(CallType.hallUp)
			clear(CallType.hallDown)
			break
		default:
			panic("Invalid direction")
		}
		break
	}
	
	return e2
} 