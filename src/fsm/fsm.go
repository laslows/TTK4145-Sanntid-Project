package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/events"
	"Sanntid/src/orders"
	"Sanntid/src/timer"
	"fmt"
	/*
		"Sanntid/src/elevator"
		"Sanntid/src/driver"
		"Sanntid/src/timer"
	*/)

//TODO: Fix naming conventions

func Fsm(e *elevator.Elevator, timetaker *timer.Timer, cabButtonCh <-chan events.ButtonEvent, floorCh <-chan int, timerCh <-chan bool, motorStopCh <-chan bool,
	assignedOrderCh <-chan orders.Order) {
	//Can only receive on channels. Might have to change tho, idk
	//Maybe make buttonevent and ordertype the samenthing
	//Putt update backup overalt lol

	for {
		select {
		case buttonEvent := <-cabButtonCh:
			NewOrder(e, buttonEvent.GetFloor(), (orders.OrderType)(buttonEvent.GetButton()), timetaker)

			fmt.Print("Received button event:", buttonEvent.GetFloor(), buttonEvent.GetButton())
		case assignedOrder := <-assignedOrderCh:
				NewOrder(e, assignedOrder.GetFloor(), assignedOrder.GetOrderType(), timetaker)
			fmt.Print("Assigned order from master: ", assignedOrder)
		case floorArrival := <-floorCh:
			onFloorArrival(e, floorArrival, timetaker)

			fmt.Print("Received floor event:", floorArrival)
		case <-timerCh:
			// Close door
			OnDoorTimeout(e, timetaker)

			fmt.Print("Received timer event")
		case <-motorStopCh:
			//Maybe make it receive a struct (MotorStopEvent, idk)

			//Inform other elevators
			//Clear queue
			//Try to reach new floor if between floors
			fmt.Print("Is motor stopped")

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

func OnDoorTimeout(e *elevator.Elevator, _timer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		pair := ChooseDirection(*e)
		e.SetDirection(pair.m_dirn)
		e.SetBehaviour(pair.m_behaviour)

		switch e.GetBehaviour() {
		case elevator.DoorOpen:
			_timer.Start(e.GetDoorOpenDuration())
			*e = ClearAtCurrentFloor(*e)
			setAllLights(*e)
		case elevator.Moving:
			fallthrough
		case elevator.Idle:
			elevator.DoorOpenLight(false)
			elevator.MotorDirection(e.GetDirection())
		}

	default:
		break
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

	setAllLights(*e)
}
