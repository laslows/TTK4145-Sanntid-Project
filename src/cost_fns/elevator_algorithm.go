package elevator

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
)

// Already have this
func requestsAbove(e Elevator) bool {
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
func requestsBelow(e Elevator) bool {
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
func AnyRequests(e Elevator) bool {
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
func AnyRequestsAtFloor(e Elevator) bool {
	for _, hasRequest := range e.GetRequests()[e.GetFloor()] {
		if hasRequest {
			return true
		}
	}
	return false
}

// TODO: Fix case-variable names and import CallType
// I think we already have this?? Maybe add checks for below 0 or higher than N_FLOORS
func ShouldStop(e Elevator) bool {
	switch e.GetDirection() {
	case Up:
		return (e.GetRequests()[e.GetFloor()][HallUp] ||
			e.GetRequests()[e.GetFloor()][Cab] ||
			!requestsAbove(e) ||
			e.GetFloor() == 0 ||
			e.GetFloor() >= len(e.GetRequests())-1)
	case Down:
		return (e.GetRequests()[e.GetFloor()][HallDown] ||
			e.GetRequests()[e.GetFloor()][Cab] ||
			!requestsBelow(e) ||
			e.GetFloor() == 0 ||
			e.GetFloor() >= len(e.GetRequests())-1)
	case Stop:
		return true
	default:
		panic("Invalid direction")
	}
}

// Already have this
func ChooseDirection(e Elevator) Direction {
	switch e.GetDirection() {
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

// I think we already have this??
func ClearReqsAtFloor(
	e *Elevator,
	clearMode config.ClearRequestType,
	onClearedRequest func(Button),
) {
	clearReq := func(c Button) {
		floor := e.GetFloor()
		btn := driver.ButtonType(c)

		if e.GetRequestAtFloor(floor, btn) {
			if onClearedRequest != nil {
				onClearedRequest(c)
			}
			e.SetRequest(floor, btn, false)
		}
	}

	switch clearMode {
	case config.All:
		for c := 0; c < config.N_BUTTONS; c++ {
			clearReq(Button(c))
		}

	case config.InDirn:
		// Always clear cab request at current floor
		clearReq(Cab)

		floor := e.GetFloor()

		switch e.GetDirection() {
		case Up:
			if e.GetRequestAtFloor(floor, driver.BT_HallUp) {
				clearReq(HallUp)
			} else if !requestsAbove(*e) {
				clearReq(HallDown)
			}

		case Down:
			if e.GetRequestAtFloor(floor, driver.BT_HallDown) {
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
