package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
)

//TODO: not pass whole elevator struct?

type dirnBehaviourPair struct {
	m_dirn      elevator.Direction
	m_behaviour elevator.ElevatorBehaviour
}

func requestsAbove(e elevator.Elevator) bool {
	for f := e.GetFloor() + 1; f < config.N_FLOORS; f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			if e.GetLocalRequestAtFloor(f, (driver.ButtonType)(btn)) {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	for f := 0; f < e.GetFloor(); f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			if e.GetLocalRequestAtFloor(f, (driver.ButtonType)(btn)) {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elevator.Elevator) bool {
	for btn := 0; btn < config.N_BUTTONS; btn++ {
		if e.GetLocalRequestAtFloor(e.GetFloor(), (driver.ButtonType)(btn)) {
			return true
		}
	}
	return false
}

func anyRequests(e elevator.Elevator) bool {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for _, hasRequest := range e.GetLocalRequests()[floor] {
			if hasRequest {
				return true
			}
		}
	}
	return false
}

func clearAtCurrentFloor(e elevator.Elevator) (elevator.Elevator, []orders.Order) {
	var completedHallOrders []orders.Order
	floor := e.GetFloor()
	hadHallUp := e.GetLocalRequestAtFloor(floor, driver.BT_HallUp)
	hadHallDown := e.GetLocalRequestAtFloor(floor, driver.BT_HallDown)

	e.SetLocalRequest(floor, driver.BT_Cab, false)

	switch e.GetDirection() {
	case elevator.Up:
		if !requestsAbove(e) && !hadHallUp {
			e.SetLocalRequest(floor, driver.BT_HallDown, false)
			if hadHallDown {
				completedHallOrders = append(completedHallOrders, orders.New(floor, orders.HALL_DOWN))
			}
		}
		e.SetLocalRequest(floor, driver.BT_HallUp, false)
		if hadHallUp {
			completedHallOrders = append(completedHallOrders, orders.New(floor, orders.HALL_UP))
		}

	case elevator.Down:
		if !requestsBelow(e) && !hadHallDown {
			e.SetLocalRequest(floor, driver.BT_HallUp, false)
			if hadHallUp {
				completedHallOrders = append(completedHallOrders, orders.New(floor, orders.HALL_UP))
			}
		}
		e.SetLocalRequest(floor, driver.BT_HallDown, false)
		if hadHallDown {
			completedHallOrders = append(completedHallOrders, orders.New(floor, orders.HALL_DOWN))
		}

	default:
		e.SetLocalRequest(floor, driver.BT_HallUp, false)
		e.SetLocalRequest(floor, driver.BT_HallDown, false)

		if hadHallUp {
			completedHallOrders = append(completedHallOrders, orders.New(floor, orders.HALL_UP))
		}
		if hadHallDown {
			completedHallOrders = append(completedHallOrders, orders.New(floor, orders.HALL_DOWN))
		}
	}

	return e, completedHallOrders
}

func shouldStop(e elevator.Elevator) bool {
	if e.GetFloor() == 0 || e.GetFloor() == config.N_FLOORS-1 {
		return true
	}

	switch e.GetDirection() {
	case elevator.Down:
		return (e.GetLocalRequestAtFloor(e.GetFloor(), driver.BT_HallDown) ||
			e.GetLocalRequestAtFloor(e.GetFloor(), driver.BT_Cab) ||
			!requestsBelow(e))
	case elevator.Up:
		return (e.GetLocalRequestAtFloor(e.GetFloor(), driver.BT_HallUp) ||
			e.GetLocalRequestAtFloor(e.GetFloor(), driver.BT_Cab) ||
			!requestsAbove(e))
	default:
		return true
	}
}

func shouldClearImmediately(e elevator.Elevator, btn_floor int, order_type orders.OrderType) bool {
	return (e.GetFloor() == btn_floor &&
		((e.GetDirection() == elevator.Up && order_type == orders.HALL_UP) ||
			(e.GetDirection() == elevator.Down && order_type == orders.HALL_DOWN) ||
			e.GetDirection() == elevator.Stop ||
			order_type == orders.CAB))
}

func chooseDirection(e elevator.Elevator) dirnBehaviourPair {
	switch e.GetDirection() {
	case elevator.Up:
		if requestsAbove(e) {
			return dirnBehaviourPair{elevator.Up, elevator.Moving}
		} else if requestsHere(e) {
			return dirnBehaviourPair{elevator.Down, elevator.DoorOpen}
		} else if requestsBelow(e) {
			return dirnBehaviourPair{elevator.Down, elevator.Moving}
		}
		return dirnBehaviourPair{elevator.Stop, elevator.Idle}
	case elevator.Down:
		if requestsBelow(e) {
			return dirnBehaviourPair{elevator.Down, elevator.Moving}
		} else if requestsHere(e) {
			return dirnBehaviourPair{elevator.Up, elevator.DoorOpen}
		} else if requestsAbove(e) {
			return dirnBehaviourPair{elevator.Up, elevator.Moving}
		}
		return dirnBehaviourPair{elevator.Stop, elevator.Idle}
	case elevator.Stop:
		if requestsHere(e) {
			return dirnBehaviourPair{elevator.Stop, elevator.DoorOpen}
		}
		if requestsAbove(e) {
			return dirnBehaviourPair{elevator.Up, elevator.Moving}
		} else if requestsBelow(e) {
			return dirnBehaviourPair{elevator.Down, elevator.Moving}
		}
	default:
		return dirnBehaviourPair{elevator.Stop, elevator.Idle}
	}

	return dirnBehaviourPair{elevator.Stop, elevator.Idle}
}
