package main

import (
	"Sanntid/config"
	"Sanntid/elevator"
	"Sanntid/elevio"
	"Sanntid/fsm"
	"Sanntid/timer"
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

		{ //Request button
			var prev [config.N_FLOORS][config.N_BUTTONS]bool

			for f := 0; f < config.N_FLOORS; f++ {
				for btn := 0; btn < config.N_BUTTONS; btn++ {
					v := elevator.RequestButton(f, (elevio.ButtonType)(btn))
					if v && v != prev[f][btn] {
						fsm.OnRequestButtonPress(elev, f, (elevator.Button)(btn), timetaker)
					}
					prev[f][btn] = v
				}
			}
		}

		{ //Floor sensor
			prev := -1
			f := elevator.FloorSensor()
			if f != -1 && f != prev {
				fsm.OnFloorArrival(elev, f, timetaker)
			}
			prev = f
		}

		//Timer
		if timetaker.TimedOut() {
			timetaker.Stop()
			fsm.OnDoorTimeout(elev, timetaker)
		}
	}

}
