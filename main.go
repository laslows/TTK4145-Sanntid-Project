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

	//Run program with "go run main.go -port=##### (default is 15657)"
	//Run simulator with "./SimElevatorServer --port #####"
	elevatorPort := flag.String("port", "15657", "port number of the elevator server")
	flag.Parse()

	driver.Init("localhost:"+*elevatorPort, config.N_FLOORS)

	elev := elevator.New(*elevatorPort)
	timetaker := timer.New()

	cabButtonCh := make(chan orders.Order)
	hallButtonCh := make(chan orders.Order)
	assignedOrdersFromMasterCh := make(chan [config.N_FLOORS][config.N_BUTTONS - 1]bool)
	localAssignedHallOrdersCh := make(chan [config.N_FLOORS][config.N_BUTTONS - 1]bool)
	floorCh := make(chan int)
	timerCh := make(chan bool)
	motorStopCh := make(chan bool)
	obstructionCh := make(chan bool)
	networkDisconnectCh := make(chan bool)
	updateWorldViewCh := make(chan elevator.Backup)
	peerLostCh := make(chan int)
	peerConnectedCh := make(chan int)

	initialize.Initialize(elev)

	go fsm.Fsm(elev, timetaker, cabButtonCh, floorCh, timerCh, motorStopCh, obstructionCh, networkDisconnectCh, localAssignedHallOrdersCh, updateWorldViewCh)
	go fsm.MasterFsm(elev, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, updateWorldViewCh, peerLostCh, peerConnectedCh)
	go events.InputPoller(cabButtonCh, hallButtonCh, floorCh, timerCh, motorStopCh, obstructionCh, networkDisconnectCh, elev, timetaker)
	go network.ListenForHeartbeats(elev, updateWorldViewCh, peerLostCh)
	go network.BroadcastHeartbeat(elev)
	go network.ListenForMessages(elev, hallButtonCh, assignedOrdersFromMasterCh, peerConnectedCh)

	select {
	// Keep main alive
	}
}
