package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

//TODO: make master redistribute orders when new peer

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrdersFromMasterCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, mergeOrdersOnBroadcastTimeoutCh chan [config.N_FLOORS][config.N_BUTTONS - 1]bool, 
	tryUpdateWorldViewCh <-chan elevator.Backup, requestRedistributionCh <-chan struct{}, peerLostCh <-chan int, peerConnectedCh <-chan int) {

	if !e.GetIsMaster() {
		fmt.Println("Immediately switching to slave")
		go slaveFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh, tryUpdateWorldViewCh, requestRedistributionCh, peerLostCh, peerConnectedCh)
		return
	}

	redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
	e.ClearDisconnectedNodeQueue()

	printOrders(e)

Loop:
	for {
		select {
		case hallOrder := <-hallButtonCh:

			if checkNewOrder(e, hallOrder) {
				fmt.Printf("New order received!")
				redistributeHallOrders(e, &hallOrder, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
			}

		case incomingOrders := <- mergeOrdersOnBroadcastTimeoutCh:

			localAssignedHallOrdersCh <- mergeHallOrders(*e, incomingOrders)

		case heartBeat := <-tryUpdateWorldViewCh:

			//TODO: maybe fix that disonnected node gets overwritten with zero when connecting again

			if !e.TryUpdateWorldView(&heartBeat) {
				continue
			}

			if e.ShouldRedistributeOrders(&heartBeat) {
				e.UpdateWorldView(&heartBeat)
				redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
				fmt.Println("Redistributing orders because of change in obstruction or motorstop status")
			} else {
				e.UpdateWorldView(&heartBeat)
			}

			onUpdateWorldView(e)

			if !e.GetIsMaster() {
				fmt.Println("Switching to slave")
				break Loop
			}

		case <-requestRedistributionCh:

			redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
			fmt.Println("Redistributing orders because of motorstop or obstruction")

		case peer := <-peerLostCh:
			fmt.Println("We lost peer ", peer)

			e.LoseConnectionToPeer(peer)
			redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
			e.ClearDisconnectedNodeQueue()

		case peer := <-peerConnectedCh:
			fmt.Println("Gained connection to peer. Sending worldview")
			network.SendWorldView(e.GetWorldView(), e.GetID(), peer)
			//redistributeHallOrders(e, nil, localAssignedHallOrdersCh)

			e.TryUpdateIsMaster()
			if !e.GetIsMaster() {
				fmt.Println("gained connection to new master, switching to slave")
				break Loop
			}

		}

		time.Sleep(10 * time.Millisecond)
	}

	//Maybe make onMasterSlaveChange-function
	e.SetIsMaster(false)
	e.UpdateMyBackupAndWorldView()
	go slaveFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh, 
		tryUpdateWorldViewCh, requestRedistributionCh, peerLostCh, peerConnectedCh)

}

func slaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrdersFromMasterCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, mergeOrdersOnBroadcastTimeoutCh chan [config.N_FLOORS][config.N_BUTTONS - 1]bool, 
	tryUpdateWorldViewCh <-chan elevator.Backup, requestRedistributionCh <-chan struct{}, peerLostCh <-chan int, peerConnectedCh <-chan int) {

	fmt.Println("I am slave")
	fmt.Printf("Master is: %d \n", e.GetMasterID())

Loop:
	for {

		select {
		case buttonEvent := <-hallButtonCh:

			network.SendHallOrder(buttonEvent, e.GetID(), e.GetMasterID())

		case incomingOrders := <- mergeOrdersOnBroadcastTimeoutCh:

			localAssignedHallOrdersCh <- mergeHallOrders(*e, incomingOrders)

		case heartBeat := <-tryUpdateWorldViewCh:

			if !e.TryUpdateWorldView(&heartBeat) {
				continue
			}

			e.UpdateWorldView(&heartBeat)
			onUpdateWorldView(e)

			if e.GetIsMaster() {
				fmt.Println("Switching to master")
				break Loop
			}

		case peer := <-peerLostCh:

			e.LoseConnectionToPeer(peer)

			e.TryUpdateIsMaster()
			if e.GetIsMaster() {
				fmt.Println("Switching to master because I lost connection to master")
				break Loop
			} else {
				e.ClearDisconnectedNodeQueue()
			}

		case <-peerConnectedCh:
			continue

		case orderList := <-assignedOrdersFromMasterCh:

			localAssignedHallOrdersCh <- orderList

		}

		time.Sleep(10 * time.Millisecond)

	}

	e.SetIsMaster(true)
	e.UpdateMyBackupAndWorldView()
	go MasterFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh, 
		tryUpdateWorldViewCh, requestRedistributionCh, peerLostCh, peerConnectedCh)

}

func onUpdateWorldView(e *elevator.Elevator) {

	e.TryUpdateIsMaster()
	setAllLights(*e)

}

func redistributeHallOrders(e *elevator.Elevator, hallOrder *orders.Order, localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, 
	mergeOrdersOnBroadcastTimeoutCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool) {

	globalOrderAssignments := runHallRequestAlgorithm(e, hallOrder)
	localAssignedHallOrdersCh <- globalOrderAssignments[e.GetID()]
	for id, orderList := range globalOrderAssignments {
		if id != e.GetID() {
			network.SendHallOrderRedistribution(orderList, e.GetID(), id, mergeOrdersOnBroadcastTimeoutCh)

			e.OverwriteHallRequestsInMasterWorldview(id, orderList)
		}
	}

	fmt.Println("Redistributed orders: ", globalOrderAssignments)

}

func mergeHallOrders(e elevator.Elevator, incomingOrderList [config.N_FLOORS][config.N_BUTTONS-1]bool) [config.N_FLOORS][config.N_BUTTONS - 1]bool {
	currentRequests := e.GetRequests()
	mergedHallRequests := [config.N_FLOORS][config.N_BUTTONS - 1]bool{}

	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS-1; button++ {
			mergedHallRequests[floor][button] = currentRequests[floor][button] || incomingOrderList[floor][button]
		}
	}

	return mergedHallRequests
}

// TODO: delete
func printOrders(e *elevator.Elevator) {
	//Print all known orders in worldview as [][]

	hallRequests := [config.N_FLOORS][config.N_BUTTONS - 1]bool{}

	for _, backup := range e.GetWorldView() {
		if backup == nil {
			continue
		}

		backupRequests := backup.GetRequests()
		for i, row := range backupRequests {
			for j, value := range row[:len(row)-1] {
				hallRequests[i][j] = hallRequests[i][j] || value
			}
		}
	}

	fmt.Println("Known orders in worldview: ", hallRequests)
}
