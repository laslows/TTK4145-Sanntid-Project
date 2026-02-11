package fsm

import(
	"elevator"
	"requests"
	"timer"
)

func setAllLights(Elevator e){
	for(floor := 0; floor < elevator.N_FLOORS; floor++){
		for(btn := 0; btn < elevator.N_BUTTONS; btn++){
			elevio.RequestButtonLight(btn, floor, e.GetRequest(floor, btn))
		}
	}
}

func onInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevator.Down)
	e.SetDirection(elevator.Down)
	e.SetBehaviour(elevator.Moving)
}

func onRequestButtonPress(e *elevator.Elevator, floor int, button elevator.Button) {

	switch(e.GetBehaviour()){
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(e, floor, button) {
			timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(floor, button, true)
		}
		break;
	case elevator.Moving:
		e.SetRequest(floor, button, true)
		break;
	case elevator.Idle:
		e.SetRequest(floor, button, true)
		pair := requests.ChooseDrection(*e)
		e.SetDirection(pair.GetDirection())
		e.SetBehaviour(pair.GetBehaviour())

		switch(pair.GetDirection()){
		case elevator.DoorOpen:
			elevio.DoorOpenLight(true)
			timer.Start(e.GetDoorOpenDuration())
			*e = requests.ClearAtCurrentFloor(*e)
			break;

		case elevator.Moving:
			elevator.MotorDirection(pair.GetDirection())
			break;

		case elevator.Idle:
			break;

		break;
	}

	setAllLights(*e)

}


func onFloorArrival(e *elevator.Elevator, floor int){

	e.SetFloor(floor)
	elevator.FloorIndicator(floor)

	switch(e.GetBehaviour()){
	case elevator.Moving:
		if requests.ShouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = requests.ClearAtCurrentFloor(*e)
			timer.Start(e.GetDoorOpenDuration())
			setAllLights(*e)
			e.SetBehaviour(elevator.DoorOpen)
		}
		break;

	default:
		break;
	}
}

func onDoorTimeout(e *elevator.Elevator) {

}