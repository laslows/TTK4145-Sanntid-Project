package main

import (
	"flag"

	"Sanntid/src/config"
	"Sanntid/src/driver"
	"Sanntid/src/elevator"
	"Sanntid/src/events"
	"Sanntid/src/fsm"
	"Sanntid/src/initialize"
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"Sanntid/src/timer"
)

func main() {

	//Run program with "go run main.go -port=#####
	elevatorPort := flag.String("port", "15657", "port number of the elevator server")
	flag.Parse()

	driver.Init("localhost:"+*elevatorPort, config.N_FLOORS)

	elev := elevator.New(*elevatorPort)
	timetaker := timer.New()

	cabButtonCh := make(chan events.ButtonEvent)
	hallButtonCh := make(chan events.ButtonEvent)
	assignedOrderCh := make(chan orders.Order)
	floorCh := make(chan int)
	timerCh := make(chan bool)
	motorStopCh := make(chan bool)

	initialize.Initialize(elev)

	go fsm.Fsm(elev, timetaker, cabButtonCh, floorCh, timerCh, motorStopCh, assignedOrderCh)
	go fsm.MasterSlaveFsm(hallButtonCh, assignedOrderCh)
	go events.InputPoller(cabButtonCh, hallButtonCh, floorCh, timerCh, motorStopCh, elev, timetaker)
	go network.ListenForHeartbeats(elev)
	go network.BroadcastHeartbeat(elev)

	for {
		// Keep main goroutine alive
	}
}
