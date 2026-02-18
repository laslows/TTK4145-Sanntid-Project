package fsm

import (
	elevio "../driver-go/elevio"
	elevator "../elevator"
	requests "../requests"
	timer "../timer"
)

func SetAllLights(e *elevator.Elevator) {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			_ = e.GetRequest(floor, elevator.ButtonType(btn))
		}
	}
}

func OnInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.SetDirection(elevator.Down)
	e.SetBehaviour(elevator.Moving)
}

func OnRequestButtonPress(e *elevator.Elevator, floor int, button elevator.ButtonType) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(*e, floor, button) {
			t := timer.New()
			t.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(floor, button, true)
		}
	case elevator.Moving:
		e.SetRequest(floor, button, true)
	case elevator.Idle:
		e.SetRequest(floor, button, true)
		pair := requests.ChooseDirection(*e)
		e.SetDirection(pair.GetDirection())
		e.SetBehaviour(pair.GetBehaviour())

		switch pair.GetDirection() {
		case elevator.Down:
		case elevator.Stop:
		case elevator.Up:
		}
	}
}

func OnFloorArrival(e *elevator.Elevator, floor int) {

}

func OnDoorTimeout(e *elevator.Elevator) {

}

func FSM() {
	// TODO
}
