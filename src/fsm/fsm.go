package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/events"
	"Sanntid/src/orders"
	"Sanntid/src/timer"
	/*
		"Sanntid/src/elevator"
		"Sanntid/src/driver"
		"Sanntid/src/timer"
	*/)

//TODO: Fix naming conventions

func Fsm(buttonCh <-chan events.ButtonEvent, floorCh <-chan int, timerCh <-chan bool, motorStopCh <-chan bool,
	assignedOrderChan <-chan orders.OrderType) {
	//Can only receive on channels. Might have to change tho, idk
	//Maybe make buttonevent and ordertype the samenthing

	for {
		select {
		case buttonEvent := <-buttonCh:
			if buttonEvent.GetButton() == elevator.Cab {
				//Put in queue
			} else {
				//Tell master
			}

			println("Received button event:", buttonEvent.GetFloor(), buttonEvent.GetButton())

		case floorArrival := <-floorCh:

			println("Received floor event:", floorArrival)
		case <-timerCh:
			// Close door
			println("Received timer event")
		case <-motorStopCh:
			//Maybe make it receive a struct (MotorStopEvent, idk)

			//Inform other elevators
			//Clear queue
			//Try to reach new floor if between floors
			println("Is motor stopped")
		case assignedOrder := <-assignedOrderChan:

			println("Assigned order from master: ", assignedOrder)
		}

	}

}

func onFloorArrival(e *elevator.Elevator, floor int, _timer *timer.Timer) {
	// Clear floor from queue
	// Tell network
	// Stop motor

	e.SetFloor(floor)
	elevator.FloorIndicator(floor)

	switch e.GetBehaviour() {
	case elevator.Moving:
		if ShouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = ClearAtCurrentFloor(*e)
			_timer.Start(e.GetDoorOpenDuration())
			//setAllLights(*e)
			e.SetBehaviour(elevator.DoorOpen)
		}

	default:
		break
	}
}

func setAllLights(e elevator.Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (driver.ButtonType)(btn), e.GetRequestAtFloor(floor, (driver.ButtonType)(btn)))
		}
	}
}

/*
func OnRequestButtonPress(e *elevator.Elevator, floor int, button elevator.Button, _timer *timer.Timer) {

	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(*e, floor, button) {
			_timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(floor, (driver.ButtonType)(button), true)
		}

	case elevator.Moving:
		e.SetRequest(floor, (driver.ButtonType)(button), true)

	case elevator.Idle:
		e.SetRequest(floor, (driver.ButtonType)(button), true)
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

	}

	//setAllLights(*e)
}
*/
