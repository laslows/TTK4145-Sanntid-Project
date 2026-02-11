package main

import (
	//"Sanntid/config"
	"Sanntid/config"
	"Sanntid/fsm"
	"Sanntid/timer"

	//"Sanntid/timer"
	"Sanntid/elevator"
	"Sanntid/elevio"
)

func main() {

	elevio.Init("localhost:15657", config.N_FLOORS)

	elev := elevator.New()
	timetaker := timer.New()
	//inputPollRate := 25

	if elevator.FloorSensor() == -1 {
		fsm.OnInitBetweenFloors(elev)
	}

	for {
		if timetaker.TimedOut() {
			timetaker.Stop()
			fsm.OnDoorTimeout(elev, timetaker)
		}
	}

}
