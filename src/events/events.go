package events

import (
	"fmt"
	"time"

	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"Sanntid/src/timer"
)

func InputPoller(cabButtonCh chan<- orders.Order, hallButtonCh chan<- orders.Order, floorCh chan<- int,
	timerCh chan<- bool, motorStopCh chan<- bool, obstructionCh chan<- bool, networkDisconnectCh chan<- bool, e *elevator.Elevator, timetaker *timer.Timer) {
	//Can only send on channels, not receive.

	var prevButtons [config.N_FLOORS][config.N_BUTTONS]bool
	var prevFloor int = -1
	var prevStopButton bool = false

	motorStopWatchdog := timer.New()
	var requestResetWatchdog = true

	var obstruction = false

	for {
		// Check new button pressed
		for f := 0; f < config.N_FLOORS; f++ {
			for btn := 0; btn < config.N_BUTTONS; btn++ {
				v := elevator.RequestButton(f, (driver.ButtonType)(btn))
				if v && v != prevButtons[f][btn] {

					if btn == int(elevator.Cab) {
						cabButtonCh <- orders.New(f, (orders.OrderType)(btn))
					} else {
						hallButtonCh <- orders.New(f, (orders.OrderType)(btn))
					}
					//Should trigger OnRequestButtonPress
					//fmt.Printf("Button %d on floor %d pressed\n", btn, f)
				}
				prevButtons[f][btn] = v
			}
		}

		// Check stop button for network disconnect
		stopPressed := elevator.StopButton()
		if stopPressed && !prevStopButton {
			fmt.Println("Stop button pressed, toggling network")
			networkDisconnectCh <- true
		}
		prevStopButton = stopPressed

		//Check new floor arrival
		f := elevator.FloorSensor()
		if f != -1 && f != prevFloor {
			floorCh <- f

			requestResetWatchdog = true
			motorStopWatchdog.Stop()

			//Should trigger OnFloorArrival
		} else if f == -1 && requestResetWatchdog {
			motorStopWatchdog.Start(time.Duration(config.MOTOR_STOP_TIMEOUT) * time.Second)
			requestResetWatchdog = false
		}
		prevFloor = f

		if motorStopWatchdog.TimedOut() {
			motorStopWatchdog.Stop()
			motorStopCh <- true
		}

		if timetaker.TimedOut() {
			timetaker.Stop()
			//Can make timerEvent struct if we later find out that we need it
			timerCh <- true
		}

		if e.GetBehaviour() == elevator.DoorOpen && elevator.ObstructionSwitch() {
			timetaker.Start(e.GetDoorOpenDuration())

			if !obstruction {
				obstruction = true
				obstructionCh <- true
			}
		} else if obstruction && !elevator.ObstructionSwitch() {
			obstruction = false
			obstructionCh <- false
		}

		time.Sleep(time.Duration(config.INPUT_POLL_RATE) * time.Millisecond)
	}
}
