package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
)

func requestsAbove(e elevator.Elevator) bool {
	for f := e.GetFloor() + 1; f < config.N_FLOORS; f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			if e.GetRequestAtFloor(f, (driver.ButtonType)(btn)) {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	for f := 0; f < e.GetFloor(); f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			if e.GetRequestAtFloor(f, (driver.ButtonType)(btn)) {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elevator.Elevator) bool {
	for btn := 0; btn < config.N_BUTTONS; btn++ {
		if e.GetRequestAtFloor(e.GetFloor(), (driver.ButtonType)(btn)) {
			return true
		}
	}
	return false
}

func ClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {

	e.SetRequest(e.GetFloor(), driver.BT_Cab, false)

	switch e.GetDirection() {
	case elevator.Up:
		if !requestsAbove(e) && !e.GetRequestAtFloor(e.GetFloor(), driver.BT_HallUp) {
			e.SetRequest(e.GetFloor(), driver.BT_HallDown, false)
		}
		e.SetRequest(e.GetFloor(), driver.BT_HallUp, false)

	case elevator.Down:
		if !requestsBelow(e) && !e.GetRequestAtFloor(e.GetFloor(), driver.BT_HallDown) {
			e.SetRequest(e.GetFloor(), driver.BT_HallUp, false)
		}
		e.SetRequest(e.GetFloor(), driver.BT_HallDown, false)

	default:
		e.SetRequest(e.GetFloor(), driver.BT_HallUp, false)
		e.SetRequest(e.GetFloor(), driver.BT_HallDown, false)
	}
	return e
}

func ShouldStop(e elevator.Elevator) bool {
	switch e.GetDirection() {
	case elevator.Down:
		return (e.GetRequestAtFloor(e.GetFloor(), driver.BT_HallDown) ||
			e.GetRequestAtFloor(e.GetFloor(), driver.BT_Cab) ||
			!requestsBelow(e))
	case elevator.Up:
		return (e.GetRequestAtFloor(e.GetFloor(), driver.BT_HallUp) ||
			e.GetRequestAtFloor(e.GetFloor(), driver.BT_Cab) ||
			!requestsAbove(e))
	default:
		return true
	}
}
