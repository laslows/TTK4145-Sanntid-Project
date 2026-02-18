package requests

import (
	elevator "../elevator"
)

type Elevator = elevator.Elevator

type DirnBehaviourPair struct {
	m_dirn      elevator.Direction
	m_behaviour elevator.ElevatorBehaviour
}

func (p DirnBehaviourPair) GetDirection() elevator.Direction {
	return p.m_dirn
}

func (p DirnBehaviourPair) GetBehaviour() elevator.ElevatorBehaviour {
	return p.m_behaviour
}

// Hjelpefunksjoner.
func requestsAbove(e Elevator) bool {
	for f := e.GetFloor() + 1; f < elevator.N_FLOORS; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			if e.GetRequest(f, elevator.ButtonType(btn)) {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e Elevator) bool {
	for f := 0; f < e.GetFloor(); f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			if e.GetRequest(f, elevator.ButtonType(btn)) {
				return true
			}
		}
	}
	return false
}

func requestsHere(e Elevator) bool {
	for btn := 0; btn < elevator.N_BUTTONS; btn++ {
		if e.GetRequest(e.GetFloor(), elevator.ButtonType(btn)) {
			return true
		}
	}
	return false
}

func ChooseDirection(e Elevator) DirnBehaviourPair {
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
		return DirnBehaviourPair{elevator.Stop, elevator.Idle}
	}
	return DirnBehaviourPair{elevator.Stop, elevator.Idle}
}

func ShouldStop(e Elevator) bool {
	switch e.GetDirection() {
	case elevator.Down:
		return (e.GetRequest(e.GetFloor(), elevator.HallDown) ||
			e.GetRequest(e.GetFloor(), elevator.Cab) ||
			(!requestsBelow(e)))
	case elevator.Up:
		return (e.GetRequest(e.GetFloor(), elevator.HallUp) ||
			e.GetRequest(e.GetFloor(), elevator.Cab) ||
			(!requestsAbove(e)))
	default:
		return true
	}
}

func ShouldClearImmediately(e Elevator, btn_floor int, btn_type elevator.ButtonType) bool {
	return (e.GetFloor() == btn_floor &&
		((e.GetDirection() == elevator.Up && btn_type == elevator.HallUp) ||
			(e.GetDirection() == elevator.Down && btn_type == elevator.HallDown) ||
			e.GetDirection() == elevator.Stop ||
			(btn_type == elevator.Cab)))
}

func ClearAtCurrentFloor(e Elevator) Elevator {
	e.SetRequest(e.GetFloor(), elevator.Cab, false)
	switch e.GetDirection() {
	case elevator.Up:
		if !requestsAbove(e) && !e.GetRequest(e.GetFloor(), elevator.HallUp) {
			e.SetRequest(e.GetFloor(), elevator.HallDown, false)
		}
		e.SetRequest(e.GetFloor(), elevator.HallUp, false)
	case elevator.Down:
		if !requestsBelow(e) && !e.GetRequest(e.GetFloor(), elevator.HallDown) {
			e.SetRequest(e.GetFloor(), elevator.HallUp, false)
		}
		e.SetRequest(e.GetFloor(), elevator.HallDown, false)
	default:
		e.SetRequest(e.GetFloor(), elevator.HallUp, false)
		e.SetRequest(e.GetFloor(), elevator.HallDown, false)
	}
	return e
}
