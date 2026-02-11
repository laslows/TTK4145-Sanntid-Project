package main

import (
	"fmt"

	"./src/fsm"
	initialize "./src/init"
	"./src/network"
	"./src/timer"
)

func main() {
	fmt.Printf("Started!\n")

	//Elevator elevator = elevator_uninitialized()
	//inputPollRate_ms int = 25

	initialize.Initialize()
	// -- Code Under --
	/* LoadConfig("elevator.con",
		ConVal("doorOpenDuration_s", "%d", &elevator.doorOpenDuration_s),
		ConVal("floorTravelDuration_s", "%d", &elevator.floorTravelDuration_s),
		ConVal("inputPollRate_ms", "%d", &elevator.inputPollRate_ms),
	)

	if(Elevator_floorSensor() == -1) {
		Fsm_onInitBetweenFloors(&elevator);
	}

	+ BUTTON_HANDLER?
	*/

	go fsm.Fsm()
	go network.Network()
	go timer.Timer()
}
