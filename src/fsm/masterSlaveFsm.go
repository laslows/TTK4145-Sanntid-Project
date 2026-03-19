package fsm

import (
	"fmt"
	"time"
	
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"Sanntid/src/orders"
)

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrdersFromMasterCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, mergeOrdersOnBroadcastTimeoutCh chan [config.N_FLOORS][config.N_BUTTONS - 1]bool, 
	tryUpdateWorldViewCh <-chan elevator.Backup, requestRedistributionCh <-chan struct{}, peerLostCh <-chan int, peerConnectedCh <-chan int) {

	if !e.GetIsMaster() {
		go slaveFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh, tryUpdateWorldViewCh, requestRedistributionCh, peerLostCh, peerConnectedCh)
		return
	}

	redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
	e.ClearDisconnectedNodeQueue()

Loop:
	for {
		select {
		case hallOrder := <-hallButtonCh:
			if checkNewOrder(e, hallOrder) {
				redistributeHallOrders(e, &hallOrder, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
			}

		case incomingOrders := <- mergeOrdersOnBroadcastTimeoutCh:
			localAssignedHallOrdersCh <- mergeHallOrders(*e, incomingOrders)

		case heartBeat := <-tryUpdateWorldViewCh:
			if !e.TryUpdateWorldView(&heartBeat) {
				continue
			}

			if shouldRedistributeOrders(e, &heartBeat) {
				e.UpdateWorldView(&heartBeat)
				redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
			} else {
				e.UpdateWorldView(&heartBeat)
			}

			onUpdateWorldView(e)

			if !e.GetIsMaster() {
				break Loop
			}

		case <-requestRedistributionCh:
			redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)

		case peer := <-peerLostCh:
			e.LoseConnectionToPeer(peer)
			redistributeHallOrders(e, nil, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh)
			e.ClearDisconnectedNodeQueue()

		case peer := <-peerConnectedCh:
			network.SendWorldView(e.GetWorldView(), e.GetID(), peer)

			e.TryUpdateIsMaster()
			if !e.GetIsMaster() {
				break Loop
			}

		}

		time.Sleep(10 * time.Millisecond)
	}

	e.SetIsMaster(false)
	e.UpdateMyBackupAndWorldView()
	fmt.Println("Switching to slave ... ")
	go slaveFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh, 
		tryUpdateWorldViewCh, requestRedistributionCh, peerLostCh, peerConnectedCh)
}

func slaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrdersFromMasterCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, mergeOrdersOnBroadcastTimeoutCh chan [config.N_FLOORS][config.N_BUTTONS - 1]bool, 
	tryUpdateWorldViewCh <-chan elevator.Backup, requestRedistributionCh <-chan struct{}, peerLostCh <-chan int, peerConnectedCh <-chan int) {
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
				break Loop
			}

		case peer := <-peerLostCh:
			e.LoseConnectionToPeer(peer)

			e.TryUpdateIsMaster()
			if e.GetIsMaster() {
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
	fmt.Println("Switchig to master ... ")
	go MasterFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, mergeOrdersOnBroadcastTimeoutCh, 
		tryUpdateWorldViewCh, requestRedistributionCh, peerLostCh, peerConnectedCh)
}

func onUpdateWorldView(e *elevator.Elevator) {
	e.TryUpdateIsMaster()
	setAllLights(*e)
}

func shouldRedistributeOrders(e *elevator.Elevator, incomingBackup *elevator.Backup) bool {
    for _, b := range e.GetWorldView() {
		if b != nil && b.GetID() == incomingBackup.GetID() {
			return (b.GetIsObstructed() != incomingBackup.GetIsObstructed() || b.GetHasMotorstop() != incomingBackup.GetHasMotorstop())
		}
	}
	return false
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
