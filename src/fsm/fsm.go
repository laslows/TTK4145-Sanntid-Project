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

func Fsm(e *elevator.Elevator, timetaker *timer.Timer, buttonCh <-chan events.ButtonEvent, floorCh <-chan int, timerCh <-chan bool, motorStopCh <-chan bool,
	assignedOrderCh <-chan orders.OrderType) {
	//Can only receive on channels. Might have to change tho, idk
	//Maybe make buttonevent and ordertype the samenthing

	for {
		select {
		case buttonEvent := <-buttonCh:
			if buttonEvent.GetButton() == elevator.Cab {
				NewOrder(e, buttonEvent.GetFloor(), (orders.OrderType)(buttonEvent.GetButton()), timetaker)
			} else {
				//Tell master
			}

			println("Received button event:", buttonEvent.GetFloor(), buttonEvent.GetButton())
		case assignedOrder := <-assignedOrderCh:

			println("Assigned order from master: ", assignedOrder)
		case floorArrival := <-floorCh:
			onFloorArrival(e, floorArrival, timetaker)

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
			setAllLights(*e)
			e.SetBehaviour(elevator.DoorOpen)
		}

	default:
		break
	}
}

func setAllLights(e elevator.Elevator) {
	globalLights := e.GetGlobalLights()

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (driver.ButtonType)(btn), globalLights[floor][btn])
		}
	}
}


func NewOrder(e *elevator.Elevator, floor int, order_type orders.OrderType, _timer *timer.Timer) {

	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if ShouldClearImmediately(*e, floor, order_type) {
			_timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(floor, (driver.ButtonType)(order_type), true)
		}

	case elevator.Moving:
		e.SetRequest(floor, (driver.ButtonType)(order_type), true)

	case elevator.Idle:
		e.SetRequest(floor, (driver.ButtonType)(order_type), true)
		pair := ChooseDirection(*e)
		e.SetDirection(pair.m_dirn)
		e.SetBehaviour(pair.m_behaviour)

		switch pair.m_behaviour {
		case elevator.DoorOpen:
			elevator.DoorOpenLight(true)
			_timer.Start(e.GetDoorOpenDuration())
			*e = ClearAtCurrentFloor(*e)

		case elevator.Moving:
			elevator.MotorDirection(pair.m_dirn)

		case elevator.Idle:
			break

		}

	}

	//setAllLights(*e)
}

