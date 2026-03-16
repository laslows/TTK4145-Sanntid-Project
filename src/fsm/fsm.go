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

func Fsm(e *elevator.Elevator, timetaker *timer.Timer, cabButtonCh <-chan orders.Order, completeHallOrderCh chan<- orders.Order, floorCh <-chan int, timerCh <-chan bool,
	motorStopCh <-chan bool, obstructionCh <-chan bool, localAssignedHallOrdersCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool, tryUpdateWorldViewCh chan<- elevator.Backup) {

	onNewOrder(e, timetaker, completeHallOrderCh)

	for {
		select {
		case buttonEvent := <-cabButtonCh:

			insertOrder(e, buttonEvent, timetaker)
			onNewOrder(e, timetaker, completeHallOrderCh)
			fmt.Printf("New cab order: floor %d, button %d\n", buttonEvent.GetFloor(), buttonEvent.GetOrderType())

		case assignedHallOrders := <-localAssignedHallOrdersCh:

			insertAllHallOrders(e, assignedHallOrders, timetaker)
			onNewOrder(e, timetaker, completeHallOrderCh)

		case floorArrival := <-floorCh:
			onFloorArrival(e, floorArrival, timetaker, completeHallOrderCh)

		case <-timerCh:
			onDoorTimeout(e, timetaker, completeHallOrderCh)

		case <-motorStopCh:

			e.SetBehaviour(elevator.MotorStop)
			e.UpdateMyBackup()

			if e.GetIsMaster() {
				tryUpdateWorldViewCh <- *e.GetMyBackup()
			}

		case obstruction := <-obstructionCh:

			e.SetIsObstructed(obstruction)
			if obstruction {
				e.SetDirection(elevator.Stop)
			}
			e.UpdateMyBackup()

			if e.GetIsMaster() {
				tryUpdateWorldViewCh <- *e.GetMyBackup()
			}
		}

	}

}

func onFloorArrival(e *elevator.Elevator, floor int, timer *timer.Timer, completeHallOrderCh chan<- orders.Order) {

	e.SetFloor(floor)
	elevator.FloorIndicator(floor)

	var completedHallOrders []orders.Order

	switch e.GetBehaviour() {
	case elevator.MotorStop:

		if !anyRequests(*e) {
			e.SetBehaviour(elevator.Idle)
			elevator.MotorDirection(elevator.Stop)

		} else if shouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e, completedHallOrders = clearAtCurrentFloor(*e)

			for _, order := range completedHallOrders {
				completeHallOrderCh <- order
			}

			timer.Start(e.GetDoorOpenDuration())
			e.SetBehaviour(elevator.DoorOpen)
			setAllLights(*e)

		} else {
			e.SetBehaviour(elevator.Moving)
		}

		e.UpdateMyBackup()

	case elevator.Moving:
		if shouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e, completedHallOrders = clearAtCurrentFloor(*e)

			for _, order := range completedHallOrders {
				completeHallOrderCh <- order
			}

			timer.Start(e.GetDoorOpenDuration())
			e.SetBehaviour(elevator.DoorOpen)
			e.UpdateMyBackup()
			setAllLights(*e)

		}
	default:
		break
	}
}

func setAllLights(e elevator.Elevator) {
	lights := e.GetAllLights()

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (driver.ButtonType)(btn), lights[floor][btn])
		}
	}
}

func onDoorTimeout(e *elevator.Elevator, timer *timer.Timer, completeHallOrderCh chan<- orders.Order) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		pair := chooseDirection(*e)
		e.SetDirection(pair.m_dirn)
		e.SetBehaviour(pair.m_behaviour)

		switch e.GetBehaviour() {
		case elevator.DoorOpen:
			timer.Start(e.GetDoorOpenDuration())
			var completedHallOrders []orders.Order
			*e, completedHallOrders = clearAtCurrentFloor(*e)

			for _, order := range completedHallOrders {
				completeHallOrderCh <- order
			}

			e.UpdateMyBackup()
			setAllLights(*e)
		case elevator.Moving:
			fallthrough
		case elevator.Idle:
			elevator.DoorOpenLight(false)
			elevator.MotorDirection(e.GetDirection())
			e.UpdateMyBackup()
		}

	default:
		break
	}
}

func insertAllHallOrders(e *elevator.Elevator, hallOrders [config.N_FLOORS][config.N_BUTTONS - 1]bool, timer *timer.Timer) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS-1; btn++ {
			if hallOrders[floor][btn] {
				insertOrder(e, orders.New(floor, orders.OrderType(btn)), timer)
			} else {
				e.SetLocalRequest(floor, (driver.ButtonType)(btn), false)
			}
		}
	}
}

func insertOrder(e *elevator.Elevator, order orders.Order, timer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if shouldClearImmediately(*e, order.GetFloor(), order.GetOrderType()) {
			fmt.Printf("Clearing order immediately: floor %d, button %d\n", order.GetFloor(), order.GetOrderType())
			timer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetLocalRequest(order.GetFloor(), (driver.ButtonType)(order.GetOrderType()), true)
		}
	default:
		e.SetLocalRequest(order.GetFloor(), (driver.ButtonType)(order.GetOrderType()), true)
	}
	e.UpdateMyBackup()
	setAllLights(*e)
}

func onNewOrder(e *elevator.Elevator, timer *timer.Timer, completeHallOrderCh chan<- orders.Order) {
	switch e.GetBehaviour() {
	case elevator.Idle:
		pair := chooseDirection(*e)
		e.SetDirection(pair.m_dirn)
		e.SetBehaviour(pair.m_behaviour)

		switch pair.m_behaviour {
		case elevator.DoorOpen:
			elevator.DoorOpenLight(true)
			timer.Start(e.GetDoorOpenDuration())
			var completedHallOrders []orders.Order
			*e, completedHallOrders = clearAtCurrentFloor(*e)

			for _, order := range completedHallOrders {
				completeHallOrderCh <- order
			}

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
