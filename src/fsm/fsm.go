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

func Fsm(e *elevator.Elevator, doorTimer *timer.Timer, cabButtonCh <-chan orders.Order, floorCh <-chan int, doorTimeoutCh <-chan bool,
	motorStopCh <-chan bool, obstructionCh <-chan bool, localAssignedHallOrdersCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool, 
	tryUpdateWorldViewCh chan<- elevator.Backup, requestRedistributionCh chan<- struct{}) {

	onNewOrder(e, doorTimer)

	for {
		select {
		case cabOrder := <-cabButtonCh:
			insertOrder(e, cabOrder, doorTimer)
			onNewOrder(e, doorTimer)

		case assignedHallOrders := <-localAssignedHallOrdersCh:
			insertAllHallOrders(e, assignedHallOrders, doorTimer)
			onNewOrder(e, doorTimer)

		case floorArrival := <-floorCh:
			onFloorArrival(e, floorArrival, doorTimer)

		case <-doorTimeoutCh:
			onDoorTimeout(e, doorTimer)

		case <-motorStopCh:

			e.SetBehaviour(elevator.MotorStop)
			e.UpdateMyBackupAndWorldView()

			if e.GetIsMaster() {
				requestRedistributionCh <- struct{}{}
			}

		case obstruction := <-obstructionCh:

			e.SetIsObstructed(obstruction)
			if obstruction {
				e.SetDirection(elevator.Stop)
			}
			e.UpdateMyBackupAndWorldView()

			if e.GetIsMaster() {
				requestRedistributionCh <- struct{}{}
			}
		}

	}

}

func onFloorArrival(e *elevator.Elevator, floor int, doorTimer *timer.Timer) {

	e.SetFloor(floor)
	elevator.FloorIndicator(floor)

	switch e.GetBehaviour() {
	case elevator.MotorStop:

		if !anyRequests(*e) {
			e.SetBehaviour(elevator.Idle)
			elevator.MotorDirection(elevator.Stop)

		} else if shouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = clearAtCurrentFloor(*e)
			doorTimer.Start(e.GetDoorOpenDuration())
			e.SetBehaviour(elevator.DoorOpen)
			setAllLights(*e)

		} else {
			e.SetBehaviour(elevator.Moving)
		}

		e.UpdateMyBackupAndWorldView()

	case elevator.Moving:
		if shouldStop(*e) {
			elevator.MotorDirection(elevator.Stop)
			elevator.DoorOpenLight(true)
			*e = clearAtCurrentFloor(*e)
			doorTimer.Start(e.GetDoorOpenDuration())
			e.SetBehaviour(elevator.DoorOpen)
			e.UpdateMyBackupAndWorldView()
			setAllLights(*e)

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

func onDoorTimeout(e *elevator.Elevator, doorTimer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		pair := chooseDirection(*e)
		e.SetDirection(pair.m_dirn)
		e.SetBehaviour(pair.m_behaviour)

		switch e.GetBehaviour() {
		case elevator.DoorOpen:
			doorTimer.Start(e.GetDoorOpenDuration())
			*e = clearAtCurrentFloor(*e)
			e.UpdateMyBackupAndWorldView()
			setAllLights(*e)
		case elevator.Moving:
			fallthrough
		case elevator.Idle:
			elevator.DoorOpenLight(false)
			elevator.MotorDirection(e.GetDirection())
			e.UpdateMyBackupAndWorldView()
		}

	default:
		break
	}
}

func insertAllHallOrders(e *elevator.Elevator, hallOrders [config.N_FLOORS][config.N_BUTTONS - 1]bool, doorTimer *timer.Timer) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS-1; btn++ {
			if hallOrders[floor][btn] {
				insertOrder(e, orders.New(floor, orders.OrderType(btn)), doorTimer)
			} else {
				e.SetRequest(floor, (driver.ButtonType)(btn), false)
			}
		}
	}
}

func insertOrder(e *elevator.Elevator, order orders.Order, doorTimer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.DoorOpen:
		if shouldClearImmediately(*e, order.GetFloor(), order.GetOrderType()) {
			fmt.Printf("Clearing order immediately: floor %d, button %d\n", order.GetFloor(), order.GetOrderType())
			doorTimer.Start(e.GetDoorOpenDuration())
		} else {
			e.SetRequest(order.GetFloor(), (driver.ButtonType)(order.GetOrderType()), true)
		}
	default:
		e.SetRequest(order.GetFloor(), (driver.ButtonType)(order.GetOrderType()), true)
	}
	e.UpdateMyBackupAndWorldView()
	setAllLights(*e)
}

func onNewOrder(e *elevator.Elevator, doorTimer *timer.Timer) {
	switch e.GetBehaviour() {
	case elevator.Idle:
		pair := chooseDirection(*e)
		e.SetDirection(pair.m_dirn)
		e.SetBehaviour(pair.m_behaviour)

		switch pair.m_behaviour {
		case elevator.DoorOpen:
			elevator.DoorOpenLight(true)
			doorTimer.Start(e.GetDoorOpenDuration())
			*e = clearAtCurrentFloor(*e)

		case elevator.Moving:
			elevator.MotorDirection(pair.m_dirn)

		case elevator.Idle:
			break
		}

	default:
		break
	}

	e.UpdateMyBackupAndWorldView()
	setAllLights(*e)

}
