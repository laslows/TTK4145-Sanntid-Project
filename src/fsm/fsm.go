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
		case elevator.Stop:
		case elevator.Up:
		break;
	}
}


func onFloorArrival(e *elevator.Elevator, floor int){

}

func onDoorTimeout(e *elevator.Elevator) {

}