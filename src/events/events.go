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
	doorTimeoutCh chan<- bool, motorStopCh chan<- bool, obstructionCh chan<- bool, e *elevator.Elevator, doorTimer *timer.Timer) {

	var prevButtons [config.N_FLOORS][config.N_BUTTONS]bool
	var prevFloor int = -1

	motorStopWatchdog := timer.New()
	var requestResetWatchdog = true

	var obstruction = false

	for {
		for floor := 0; floor < config.N_FLOORS; floor++ {
			for btn := 0; btn < config.N_BUTTONS; btn++ {
				req := elevator.RequestButton(floor, (driver.ButtonType)(btn))
				if req && req != prevButtons[floor][btn] {

					if btn == int(driver.BT_Cab) {
						cabButtonCh <- orders.New(floor, (orders.OrderType)(btn))
					} else {
						hallButtonCh <- orders.New(floor, (orders.OrderType)(btn))
					}
				}
				prevButtons[floor][btn] = req
			}
		}

		floor := driver.GetFloor()
		if floor != -1 && floor != prevFloor {
			floorCh <- floor
		}

		//hasAnyRequests := e.GetHasAnyRequests()
		// If we are between floors or if we are on a floor and should move to another floor. Count 4 seconds
		// from when elevator.moving is set to true
		// NOt sure if this works, test tomorrow :)
		// WIll only get state elevator.moving when we have requests
		watchForMotorStop := floor == -1 || (floor != -1 && floor == prevFloor && e.GetBehaviour() == elevator.Moving /*&& hasAnyRequests*/)

		if watchForMotorStop && requestResetWatchdog {
			motorStopWatchdog.Start(time.Duration(config.MOTOR_STOP_TIMEOUT) * time.Second)
			requestResetWatchdog = false
		} else if !watchForMotorStop {
			requestResetWatchdog = true
			motorStopWatchdog.Stop()
		}

		prevFloor = floor

		if motorStopWatchdog.TimedOut() {
			motorStopWatchdog.Stop()
			fmt.Println("Motor stop timeout")
			motorStopCh <- true
		}

		if doorTimer.TimedOut() {
			doorTimer.Stop()
			doorTimeoutCh <- true
		}

		if e.GetBehaviour() == elevator.DoorOpen && elevator.ObstructionSwitch() {
			doorTimer.Start(e.GetDoorOpenDuration())

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
