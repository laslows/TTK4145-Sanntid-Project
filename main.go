package main

import (
	"fmt"
	"time"

	elevator "./src/elevator"
	fsm "./src/fsm"
	initialize "./src/init"
	"./src/network"
	timer "./src/timer"
)

// This is the main function translated from the C code. TODO: Implement elsewhere
func pollLoop(inputPollRate_ms int, elev *elevator.Elevator, tmr *timer.Timer) {
	var prevBtn [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	prevFloor := -1

	// Request Button
	for f := 0; f < elevator.N_FLOORS; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			v := elevator.RequestButton(f, elevator.ButtonType(btn))
			if v && v != prevBtn[f][btn] {
				fsm.OnRequestButtonPress(elev, f, elevator.ButtonType(btn))
			}
			prevBtn[f][btn] = v
		}
	}

	// Floor Sensor TODO
	f := elevator.FloorSensor()
	if f != -1 && f != prevFloor {
		fsm.OnFloorArrival(elev, f)
	}
	prevFloor = f

	// Door Timeout TODO
	if tmr.TimedOut() {
		tmr.Stop()
		fsm.OnDoorTimeout(elev)
	}

	time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)
}

func main() {
	fmt.Printf("Started!\n")

	initialize.Initialize()
	elev := elevator.New()
	tmr := timer.New()
	// -- Code Under --
	/* LoadConfig("elevator.con",
		ConVal("doorOpenDuration_s", "%d", &elevator.doorOpenDuration_s),
		ConVal("floorTravelDuration_s", "%d", &elevator.floorTravelDuration_s),
		ConVal("inputPollRate_ms", "%d", &elevator.inputPollRate_ms),
	)

	if(elevator.FloorSensor() == -1) {
		fsm.OnInitBetweenFloors(elev);
	}

	+ BUTTON_HANDLER?
	*/

	go fsm.FSM()
	go network.Network()
	go timer.New() //timer.Timer()?

	go pollLoop(25, elev, tmr) // function translated from main.c in given code
}
