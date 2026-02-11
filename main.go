package main

import (
	"fmt"
	"time"

	elevator "./src/elevator"
	"./src/fsm"
	initialize "./src/init"
	"./src/network"
	"./src/timer"
)

func pollLoop(inputPollRate_ms int, elev *elevator.Elevator) {
	var prevBtn [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	prevFloor := -1

	// Request Button
	for f := 0; f < elevator.N_FLOORS; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			v := elevator.RequestButton(f, elevator.ButtonType(btn))
			if v && v != prevBtn[f][btn] {
				fsm.onRequestButtonPress(elev, f, btn)
			}
			prevBtn[f][btn] = v
		}
	}

	// Floor Sensor TODO
	f := elevator.FloorSensor()
	if f != -1 && f != prevFloor {
		fsm.onFloorArrival(elev, f)
	}
	prevFloor = f

	// Door Timeout TODO
	if Timer_timedOut() {
		timer.stop()
		fsm.onDoorTimeout(elev)
	}

	time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)
}

func main() {
	fmt.Printf("Started!\n")

	initialize.Initialize()
	elev := elevator.New()
	// -- Code Under --
	/* LoadConfig("elevator.con",
		ConVal("doorOpenDuration_s", "%d", &elevator.doorOpenDuration_s),
		ConVal("floorTravelDuration_s", "%d", &elevator.floorTravelDuration_s),
		ConVal("inputPollRate_ms", "%d", &elevator.inputPollRate_ms),
	)

	if(elevator.FloorSensor() == -1) {
		Fsm_onInitBetweenFloors(elev);
	}

	+ BUTTON_HANDLER?
	*/

	go fsm.Fsm()
	go network.Network()
	go timer.Timer()

	go pollLoop(25, elev) // function translated from main.c in given code
}
