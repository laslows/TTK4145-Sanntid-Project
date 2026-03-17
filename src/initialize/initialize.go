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

	}

	fmt.Println("Floor is: ", e.GetFloor())
	fmt.Println("Direction is: ", e.GetDirection())
	initOnFloor(e)

	e.TryUpdateIsMaster()
	e.UpdateMyBackup()

}

func initOnFloor(e *elevator.Elevator) {

	//TODO: spør studass om dette er nødvendig. 
	// Problemet oppstår kun dersom heisen flyttes manuelt mens den står stille, og så drepes den..
	initDirection := (int)(e.GetDirection())
	
	if initDirection == 0 && elevator.FloorSensor() == -1 {
		initDirection = -1
	}


	for elevator.FloorSensor() == -1 {
		driver.SetMotorDirection((driver.MotorDirection)(initDirection))
		e.SetBehaviour(elevator.Moving)
	}

	driver.SetMotorDirection(driver.MD_Stop)
	e.SetBehaviour(elevator.Idle)
	e.SetDirection(elevator.Stop)
	e.SetFloor(elevator.FloorSensor())

	//TODO: Spør studass om det er mulig at heien starter med døra åpen/med obstruksjon
	//e.SetIsObstructed(driver.GetObstruction())

}

func clearAllLights() {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
			elevator.RequestButtonLight(floor, (driver.ButtonType)(btn), false)
		}
	}
}
