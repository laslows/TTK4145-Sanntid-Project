package initialize

import (
	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"fmt"
)

//Move to elevator package ?

func Initialize(e *elevator.Elevator) {

	clearAllLights()

	for elevator.FloorSensor() == -1 {
		onInitBetweenFloors(e)
	}

	driver.SetMotorDirection(driver.MD_Stop)
	e.SetBehaviour(elevator.Idle)
	e.SetDirection(elevator.Stop)
	e.SetFloor(elevator.FloorSensor())

	fmt.Println("Initialiser heisen")

	fmt.Printf("Initial floor: %d\n", e.GetFloor())

	network.SendInitializationMessage(e.GetID())

	worldView, gotWorldView := network.TryListenForWorldView()

	if gotWorldView {

		for _, b := range worldView {
			if b != nil && b.GetID() == e.GetID() {
				e.RestoreElevatorState(b)
			} else if b != nil {
				e.UpdateWorldView(b)
			}
		}

		fmt.Println(e.GetRequests())

		//setAllLights(*e)
	}

	e.TryUpdateIsMaster()
	e.UpdateMyBackup()

}

func onInitBetweenFloors(e *elevator.Elevator) {
	driver.SetMotorDirection(driver.MD_DOWN)
	e.SetDirection(elevator.Down)
	e.SetBehaviour(elevator.Moving)
}

func clearAllLights() {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (driver.ButtonType)(btn), false)
		}
	}
}
