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
//TODO: maybe use defer ?? Codequalityfix

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrdersFromMasterCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, tryUpdateWorldViewCh <-chan elevator.Backup, peerLostCh <-chan int,
	peerConnectedCh <-chan int) {

	if !e.GetIsMaster() {
		fmt.Println("Immediately switching to slave")
		go slaveFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, tryUpdateWorldViewCh, peerLostCh, peerConnectedCh)
		return
	}

	// Initialize pendingOrders from worldview (picks up orders confirmed before this master started)
	pendingOrders := initPendingOrdersFromWorldView(e)
	fmt.Println("Known orders from worldview: ", pendingOrders)

	redistributeHallOrders(e, pendingOrders, localAssignedHallOrdersCh)
	e.ClearDisconnectedNodeQueue()

Loop:
	for {
		select {
		case hallOrder := <-hallButtonCh:

			f, btn := hallOrder.GetFloor(), hallOrder.GetOrderType()
			if checkNewOrder(e, hallOrder) && !pendingOrders[f][btn] {
				fmt.Printf("New order received!")
				pendingOrders[f][btn] = true
				redistributeHallOrders(e, pendingOrders, localAssignedHallOrdersCh)
			} else {
				fmt.Println("Order already in queue, not sending to algorithm")
			}

		case heartBeat := <-tryUpdateWorldViewCh:

			//TODO: maybe fix that disonnected node gets overwritten with zero when connecting again

			if !e.TryUpdateWorldView(&heartBeat) {
				continue
			}

			if e.ShouldRedistributeOrders(&heartBeat) {
				e.UpdateWorldView(&heartBeat)
				redistributeHallOrders(e, pendingOrders, localAssignedHallOrdersCh)
			} else if heartBeat.GetID() == e.GetID() {
				e.UpdateWorldView(&heartBeat)
				//Only happens if motorstop, should maybe be moved
				redistributeHallOrders(e, pendingOrders, localAssignedHallOrdersCh)
				fmt.Println("I changed obstructionstatus or motorstopstatus")
			} else {
				e.UpdateWorldView(&heartBeat)
			}

			// Clear pending orders that worldview has now confirmed
			for f := 0; f < config.N_FLOORS; f++ {
				for btn := 0; btn < config.N_BUTTONS-1; btn++ {
					if pendingOrders[f][btn] {
						for _, b := range e.GetWorldView() {
							if b != nil && b.GetRequests()[f][btn] {
								pendingOrders[f][btn] = false
								break
							}
						}
					}
				}
			}

			onUpdateWorldView(e)

			if !e.GetIsMaster() {
				fmt.Println("Switching to slave")
				break Loop
			}

		case peer := <-peerLostCh:
			fmt.Println("We lost peer ", peer)

			e.LoseConnectionToPeer(peer)

			//Must redistribute when we lose connection
			redistributeHallOrders(e, pendingOrders, localAssignedHallOrdersCh)

			e.ClearDisconnectedNodeQueue()

		case peer := <-peerConnectedCh:
			fmt.Println("We gained peer ", peer)

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
	e.UpdateMyBackup()
	go slaveFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, tryUpdateWorldViewCh, peerLostCh, peerConnectedCh)

}

// Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
// Idk om dette er en god løsning..
func slaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrdersFromMasterCh <-chan [config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, tryUpdateWorldViewCh <-chan elevator.Backup, peerLostCh <-chan int,
	peerConnectedCh <-chan int) {

	fmt.Println("I am slave")
	fmt.Printf("Master is: %d \n", e.GetMasterID())

Loop:
	for {

		select {
		case buttonEvent := <-hallButtonCh:
			//Give to masterHallOrderRequest

			network.SendHallOrder(buttonEvent, e.GetID(), e.GetMasterID())

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
			fmt.Println("We lost peer ", peer)

			e.LoseConnectionToPeer(peer)

			e.TryUpdateIsMaster()
			if e.GetIsMaster() {
				fmt.Println("Switching to master because I lost connection to everyone")
				break Loop
			} else {
				e.ClearDisconnectedNodeQueue()
			}

		case peer := <-peerConnectedCh:
			fmt.Println("We gained peer ", peer)

			fmt.Println("Gained connection to peer. I am slave, so will not send message.")

		case orderList := <-assignedOrdersFromMasterCh:

			localAssignedHallOrdersCh <- orderList

		}

		time.Sleep(10 * time.Millisecond)

	}

	e.SetIsMaster(true)
	e.UpdateMyBackup()
	go MasterFsm(e, hallButtonCh, assignedOrdersFromMasterCh, localAssignedHallOrdersCh, tryUpdateWorldViewCh, peerLostCh, peerConnectedCh)

}

func onUpdateWorldView(e *elevator.Elevator) {

	e.TryUpdateIsMaster()
	setAllLights(*e)

	//Also check motorstop
	//Also check other stuff

}

func redistributeHallOrders(e *elevator.Elevator, pendingOrders [config.N_FLOORS][config.N_BUTTONS - 1]bool, localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool) {

	globalOrderAssignments := runHallRequestAlgorithm(e, pendingOrders)
	localAssignedHallOrdersCh <- globalOrderAssignments[e.GetID()]
	for id, orderList := range globalOrderAssignments {
		if id != e.GetID() {
			network.SendHallOrderRedistribution(orderList, e.GetID(), id)
		}
	}

}

func initPendingOrdersFromWorldView(e *elevator.Elevator) [config.N_FLOORS][config.N_BUTTONS - 1]bool {
	pendingOrders := [config.N_FLOORS][config.N_BUTTONS - 1]bool{}
	for _, backup := range e.GetWorldView() {
		if backup == nil {
			continue
		}
		requests := backup.GetRequests()
		for f := 0; f < config.N_FLOORS; f++ {
			for btn := 0; btn < config.N_BUTTONS-1; btn++ {
				pendingOrders[f][btn] = pendingOrders[f][btn] || requests[f][btn]
			}
		}
	}
	return pendingOrders
}
