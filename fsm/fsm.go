package fsm

import (
	"Sanntid/config"
	"Sanntid/elevator"
	"Sanntid/elevio"
	"Sanntid/requests"
	"Sanntid/timer"
)

func setAllLights(e elevator.Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (elevio.ButtonType)(btn), e.GetRequest(floor, (elevio.ButtonType)(btn)))
		}
	}
}

func OnInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_DOWN)
	e.SetDirection(elevator.Down)
	e.SetBehaviour(elevator.Moving)
}

func OnRequestButtonPress(e *elevator.Elevator, floor int, button elevator.Button, _timer *timer.Timer) {

	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(*e, floor, button) {
			_timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(floor, (elevio.ButtonType)(button), true)
		}

	case elevator.Moving:
		e.SetRequest(floor, (elevio.ButtonType)(button), true)

	case elevator.Idle:
		e.SetRequest(floor, (elevio.ButtonType)(button), true)
		pair := requests.ChooseDirection(*e)
		e.SetDirection(pair.GetDirection())
		e.SetBehaviour(pair.GetBehaviour())

		switch pair.GetBehaviour() {
		case elevator.DoorOpen:
			elevator.DoorOpenLight(true)
			_timer.Start(e.GetDoorOpenDuration())
			*e = requests.ClearAtCurrentFloor(*e)

		case elevator.Moving:
			elevator.MotorDirection(pair.GetDirection())

		case elevator.Idle:
			break

		}

		setAllLights(*e)

	}
}

func OnFloorArrival(e *elevator.Elevator, floor int, _timer *timer.Timer) {

	e.SetFloor(floor)
	elevator.FloorIndicator(floor)

	switch e.GetBehaviour() {
	case elevator.Moving:
		if requests.ShouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = requests.ClearAtCurrentFloor(*e)
			_timer.Start(e.GetDoorOpenDuration())
			setAllLights(*e)
			e.SetBehaviour(elevator.DoorOpen)
		}

	default:
		break
	}
}

func OnDoorTimeout(e *elevator.Elevator, _timer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		pair := requests.ChooseDirection(*e)
		e.SetDirection(pair.GetDirection())
		e.SetBehaviour(pair.GetBehaviour())

		switch e.GetBehaviour() {
		case elevator.DoorOpen:
			_timer.Start(e.GetDoorOpenDuration())
			*e = requests.ClearAtCurrentFloor(*e)
			setAllLights(*e)
		case elevator.Moving:
		case elevator.Idle:
			elevator.DoorOpenLight(false)
			elevator.MotorDirection(e.GetDirection())
		}

	default:
		break
	}
}
