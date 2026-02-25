package costfns

import (
	config "Sanntid/src/config"
	es "Sanntid/src/cost_fns" //TODO: Fix this and e2, maps without get functions?
	elevator_state "Sanntid/src/elevator"
)

// Already have this
func requestsAbove(e elevator_state.Elevator) bool {
	for floor := e.GetFloor() + 1; floor < len(e.GetRequests()); floor++ {
		for _, hasRequest := range e.GetRequests()[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}

// Already have this
func requestsBelow(e elevator_state.Elevator) bool {
	for floor := e.GetFloor() - 1; floor >= 0; floor-- {
		for _, hasRequest := range e.GetRequests()[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}

// Checks if I have any orders at any floor
func AnyRequests(e elevator_state.Elevator) bool {
	for floor := 0; floor < len(e.GetRequests()); floor++ {
		for _, hasRequest := range e.GetRequests()[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}

// Checks if elevator has any requests at current floor
func AnyRequestsAtFloor(e elevator_state.Elevator) bool {
	for _, hasRequest := range e.GetRequests()[e.GetFloor()] {
		if hasRequest {
			return true
		}
	}
	return false
}

// TODO: Fix case-variable names and import CallType
// I think we already have this?? Maybe add checks for below 0 or higher than N_FLOORS
func ShouldStop(e elevator_state.Elevator) bool {
	switch e.GetDirection() {
	case elevator_state.Up:
		return (e.GetRequests()[e.GetFloor()][HallUp] ||
			e.GetRequests()[e.GetFloor()][Cab] ||
			!requestsAbove(e) ||
			e.GetFloor() == 0 ||
			e.GetFloor() >= len(e.GetRequests())-1)
	case elevator_state.Down:
		return (e.GetRequests()[e.GetFloor()][HallDown] ||
			e.GetRequests()[e.GetFloor()][Cab] ||
			!requestsBelow(e) ||
			e.GetFloor() == 0 ||
			e.GetFloor() >= len(e.GetRequests())-1)
	case elevator_state.Stop:
		return true
	default:
		panic("Invalid direction")
	}
}

// Already have this
func ChooseDirection(e elevator_state.Elevator) Dirn {
	switch e.GetDirection() {
	case elevator_state.Up:
		if requestsAbove(e) {
			return Up
		} else if AnyRequestsAtFloor(e) {
			return Stop
		} else if requestsBelow(e) {
			return Down
		} else {
			return Stop
		}
	case elevator_state.Down, elevator_state.Stop:
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

// I think we already have this??
func ClearReqsAtFloor(
	e *elevator_state.Elevator,
	e2 es.ElevatorState,
	clearMode config.ClearRequestType,
	onClearedRequest func(CallType),
) {
	clearReq := func(c CallType) {
		if e.GetRequests()[e.GetFloor()][int(c)] {
			if onClearedRequest != nil {
				onClearedRequest(c)
			}
			e2.Requests[e2.Floor][int(c)] = false
		}
	}

	switch clearMode {
	case config.All:
		for c := 0; c < len(e2.Requests[0]); c++ {
			clearReq(CallType(c))
		}

	case config.InDirn:
		// Always clear cab request at current floor
		clearReq(Cab)

		switch e.GetDirection() {
		case elevator_state.Up:
			if e2.Requests[e2.Floor][int(HallUp)] {
				clearReq(HallUp)
			} else if !requestsAbove(*e) {
				clearReq(HallDown)
			}

		case elevator_state.Down:
			if e2.Requests[e2.Floor][int(HallDown)] {
				clearReq(HallDown)
			} else if !requestsBelow(*e) {
				clearReq(HallUp)
			}

		case elevator_state.Stop:
			clearReq(HallUp)
			clearReq(HallDown)

		default:
			panic("invalid direction")
		}

	default:
		panic("invalid clear request mode")
	}
}
