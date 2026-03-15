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

	redistributeHallOrders(e, nil, localAssignedHallOrdersCh)
	e.ClearDisconnectedNodeQueue()

	//Print all known orders in worldview as [][]

	hallRequests := [config.N_FLOORS][config.N_BUTTONS-1]bool{}

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


Loop:
	for {
		select {
		case hallOrder := <-hallButtonCh:

			if checkNewOrder(e, hallOrder) {
				fmt.Printf("New order received!")

				redistributeHallOrders(e, &hallOrder, localAssignedHallOrdersCh)

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
				redistributeHallOrders(e, nil, localAssignedHallOrdersCh)
			} else if heartBeat.GetID() == e.GetID() {
				e.UpdateWorldView(&heartBeat)
				//Only happens if motorstop, should maybe be moved
				redistributeHallOrders(e, nil, localAssignedHallOrdersCh)
				fmt.Println("I changed obstructionstatus or motorstopstatus")
			} else {
				e.UpdateWorldView(&heartBeat)
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
			redistributeHallOrders(e, nil, localAssignedHallOrdersCh)

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

func redistributeHallOrders(e *elevator.Elevator, hallOrder *orders.Order, localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool) {

	globalOrderAssignments := runHallRequestAlgorithm(e, hallOrder)
	localAssignedHallOrdersCh <- globalOrderAssignments[e.GetID()]
	for id, orderList := range globalOrderAssignments {
		if id != e.GetID() {
			network.SendHallOrderRedistribution(orderList, e.GetID(), id)
		}
	}

}
