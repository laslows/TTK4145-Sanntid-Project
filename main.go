package main

import (
	"flag"
	"fmt"

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

	elevatorPort := flag.String("port", "15657", "port number of the elevator server")
	elevatorID := flag.String("id", "", "elevator id")
	flag.Parse()

	fmt.Println(*elevatorID)

	driver.Init("localhost:"+*elevatorPort, config.N_FLOORS)

	elev := elevator.New(*elevatorID)
	doorTimer := timer.New()

	cabButtonCh := make(chan orders.Order)
	hallButtonCh := make(chan orders.Order)
	assignedOrdersFromMasterCh := make(chan [config.N_FLOORS][config.N_BUTTONS - 1]bool)
	localAssignedHallOrdersCh := make(chan [config.N_FLOORS][config.N_BUTTONS - 1]bool)
	requestRedistributionCh := make(chan struct{})
	floorCh := make(chan int)
	doorTimeoutCh := make(chan bool)
	motorStopCh := make(chan bool)
	obstructionCh := make(chan bool)
	tryUpdateWorldViewCh := make(chan elevator.Backup)
	peerLostCh := make(chan int)
	peerConnectedCh := make(chan int)

	initialize.Initialize(elev)

	go fsm.Fsm(elev, doorTimer, cabButtonCh, floorCh, doorTimeoutCh, motorStopCh, obstructionCh, localAssignedHallOrdersCh, tryUpdateWorldViewCh, requestRedistributionCh)
	go fsm.MasterFsm(elev, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, tryUpdateWorldViewCh,
		requestRedistributionCh, peerLostCh, peerConnectedCh)
	go events.InputPoller(cabButtonCh, hallButtonCh, floorCh, doorTimeoutCh, motorStopCh, obstructionCh, elev, doorTimer)
	go network.ListenForHeartbeats(tryUpdateWorldViewCh, peerLostCh)
	go network.BroadcastHeartbeat(elev)
	go network.ListenForMessages(elev, hallButtonCh, assignedOrdersFromMasterCh, peerConnectedCh)

	select {}
}
