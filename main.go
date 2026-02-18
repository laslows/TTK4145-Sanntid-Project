package main

import (
	"flag"

	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/events"
	"Sanntid/src/fsm"
	"Sanntid/src/initialize"
	"Sanntid/src/timer"
)

func main() {

	elevatorPort := flag.String("port", "15657", "port number of the elevator server")
	flag.Parse()

	driver.Init("localhost:"+*elevatorPort, config.N_FLOORS)

	elev := elevator.New(*elevatorPort)
	timetaker := timer.New()

	buttonCh := make(chan events.ButtonEvent)
	floorCh := make(chan int)
	timerCh := make(chan bool)
	motorStopCh := make(chan bool)

	initialize.Initialize(elev)

	go fsm.Fsm(elev, timetaker, buttonCh, floorCh, timerCh, motorStopCh)
	go events.InputPoller(buttonCh, floorCh, timerCh, motorStopCh, timetaker)

	for {
		// Keep main goroutine alive
	}
}
