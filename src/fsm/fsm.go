package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"Sanntid/src/timer"
	"fmt"
)

//TODO: Fix naming conventions

func Fsm(e *elevator.Elevator, timetaker *timer.Timer, cabButtonCh <-chan orders.Order, floorCh <-chan int, timerCh <-chan bool,
	motorStopCh <-chan bool, obstructionCh <-chan bool, localAssignedHallOrdersCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool) {
	//Can only receive on channels. Might have to change tho, idk
	//Maybe make buttonevent and ordertype the samenthing
	//Putt update backup overalt lol

	for {
		select {
		case buttonEvent := <-cabButtonCh:

			insertOrder(e, buttonEvent, timetaker)
			onNewOrder(e, timetaker)
			fmt.Printf("New cab order: floor %d, button %d\n", buttonEvent.GetFloor(), buttonEvent.GetOrderType())

		case assignedHallOrders := <-localAssignedHallOrdersCh:

			insertAllHallOrders(e, assignedHallOrders, timetaker)
			onNewOrder(e, timetaker)

		case floorArrival := <-floorCh:
			onFloorArrival(e, floorArrival, timetaker)

		case <-timerCh:
			// Close door
			OnDoorTimeout(e, timetaker)

		case <-motorStopCh:
			//Maybe make it receive a struct (MotorStopEvent, idk)

			//Inform other elevators
			//Clear queue
			//Try to reach new floor if between floors
			e.SetBehaviour(elevator.MotorStop)
			e.UpdateMyBackup()
			//network.SendMotorStopMessage(e.GetID(), e.GetMasterID(), true)
		case obstruction := <-obstructionCh:

			
			e.SetIsObstructed(obstruction)
			e.UpdateMyBackup()
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
	case elevator.MotorStop:
		//This might be wrong, we are cooked

		if !anyRequests(*e) {
			e.SetBehaviour(elevator.Idle)
			elevator.MotorDirection(elevator.Stop)

		} else if ShouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = ClearAtCurrentFloor(*e)
			_timer.Start(e.GetDoorOpenDuration())
			setAllLights(*e)
			e.SetBehaviour(elevator.DoorOpen)

		} else {
			e.SetBehaviour(elevator.Moving)
		}

		e.UpdateMyBackup()

		//network.SendMotorStopMessage(e.GetID(), e.GetMasterID(), false)

	case elevator.Moving:
		if ShouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = ClearAtCurrentFloor(*e)
			_timer.Start(e.GetDoorOpenDuration())
			setAllLights(*e)
			e.UpdateMyBackup()
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

	e.UpdateMyBackup()
}

func insertAllHallOrders(e *elevator.Elevator, hallOrders [config.N_FLOORS][config.N_BUTTONS - 1]bool, timer *timer.Timer) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS-1; btn++ {
			if hallOrders[floor][btn] {
				insertOrder(e, orders.New(floor, orders.OrderType(btn)), timer)
			} else {
				e.SetRequest(floor, (driver.ButtonType)(btn), false)
			}
		}
	}
}

func insertOrder(e *elevator.Elevator, order orders.Order, timer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if ShouldClearImmediately(*e, order.GetFloor(), order.GetOrderType()) {
			fmt.Printf("Clearing order immediately: floor %d, button %d\n", order.GetFloor(), order.GetOrderType())
			timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(order.GetFloor(), (driver.ButtonType)(order.GetOrderType()), true)
		}
	default:
		e.SetRequest(order.GetFloor(), (driver.ButtonType)(order.GetOrderType()), true)
	}
	e.UpdateMyBackup()
	setAllLights(*e)
}

func onNewOrder(e *elevator.Elevator, _timer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.Idle:
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

	default:
		break
	}

	e.UpdateMyBackup()
	setAllLights(*e)

}

/*
func NewOrder(e *elevator.Elevator, floor int, order_type orders.OrderType, _timer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if ShouldClearImmediately(*e, floor, order_type) {
			fmt.Printf("Clearing order immediately: floor %d, button %d\n", floor, order_type)
			_timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(floor, (driver.ButtonType)(order_type), true)
		}

	case elevator.Moving:
		e.SetRequest(floor, (driver.ButtonType)(order_type), true)

	case elevator.MotorStop:
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

	default:
		break
	}
	e.UpdateMyBackup()
	setAllLights(*e)
}

*/
