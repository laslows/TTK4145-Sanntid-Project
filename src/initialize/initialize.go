package initialize

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"fmt"
)

func Initialize(e *elevator.Elevator) {

	clearAllLights()
	elevator.DoorOpenLight(false)

	fmt.Println("Initialiser heisen")
	fmt.Printf("Initial floor: %d\n", e.GetFloor())

	network.SendInitializationMessage(e.GetID(), e)
	worldView, gotWorldView := network.TryListenForWorldView()

	if gotWorldView {

		for _, b := range worldView {
			if b != nil && b.GetID() == e.GetID() {
				e.RestoreElevatorState(b)
			} else if b != nil {
				e.UpdateWorldView(b)
			}
		}

	}

	fmt.Println("Direction is: ", e.GetDirection())
	initOnFloor(e)

	e.TryUpdateIsMaster()
	e.UpdateMyBackup()

}

func initOnFloor(e *elevator.Elevator) {

	for elevator.FloorSensor() == -1 {
		driver.SetMotorDirection((driver.MotorDirection)(e.GetDirection()))
		e.SetBehaviour(elevator.Moving)
	}

	driver.SetMotorDirection(driver.MD_Stop)
	e.SetBehaviour(elevator.Idle)
	e.SetDirection(elevator.Stop)
	e.SetFloor(elevator.FloorSensor())

}

func clearAllLights() {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (driver.ButtonType)(btn), false)
		}
	}
}
