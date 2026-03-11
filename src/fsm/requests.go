package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
)

//TODO: not pass whole elevator struct? 

type DirnBehaviourPair struct {
	m_dirn      elevator.Direction
	m_behaviour elevator.ElevatorBehaviour
}

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

func anyRequests(e elevator.Elevator) bool {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for _, hasRequest := range e.GetRequests()[floor] {
			if hasRequest {
				return true
			}
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

func ShouldClearImmediately(e elevator.Elevator, btn_floor int, order_type orders.OrderType) bool {
	return (e.GetFloor() == btn_floor &&
		((e.GetDirection() == elevator.Up && order_type == orders.HALL_UP) ||
			(e.GetDirection() == elevator.Down && order_type == orders.HALL_DOWN) ||
			e.GetDirection() == elevator.Stop ||
			order_type == orders.CAB))
}

func ChooseDirection(e elevator.Elevator) DirnBehaviourPair {
	switch e.GetDirection() {
	case elevator.Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevator.Up, elevator.Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevator.Down, elevator.DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevator.Down, elevator.Moving}
		}
		return DirnBehaviourPair{elevator.Stop, elevator.Idle}
	case elevator.Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevator.Down, elevator.Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevator.Up, elevator.DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevator.Up, elevator.Moving}
		}
		return DirnBehaviourPair{elevator.Stop, elevator.Idle}
	case elevator.Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{elevator.Stop, elevator.DoorOpen}
		}
		if requestsAbove(e) {
			return DirnBehaviourPair{elevator.Up, elevator.Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevator.Down, elevator.Moving}
		}
	default:
		return DirnBehaviourPair{elevator.Stop, elevator.Idle}
	}

	return DirnBehaviourPair{elevator.Stop, elevator.Idle}
}
