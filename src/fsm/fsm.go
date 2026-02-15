package fsm

import (
	"Sanntid/src/events"
	/*
		"Sanntid/src/elevator"
		"Sanntid/src/driver"
		"Sanntid/src/timer"
	*/)

//TODO: Fix naming conventions

func Fsm(buttonCh <-chan events.ButtonEvent, floorCh <-chan int, timerCh <-chan bool, motorStopCh <-chan bool) {
	//Can only receive on channels. Might have to change tho, idk

	for {
		select {
		case buttonEvent := <-buttonCh:
			// If cab order, put in queue
			// If hall order, tell master
			println("Received button event:", buttonEvent.GetFloor(), buttonEvent.GetButton())
		case floorArrival := <-floorCh:
			// Clear floor from queue
			// Tell network
			// Stop motor
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
