package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

//TODO:
// Insert hall order to queue how to do this?
// Fix newOrder

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, globalAssignedHallOrdersCh <-chan map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, updateWorldViewCh <-chan elevator.Backup, peerLostCh <-chan int) {
Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:

			if checkNewOrder(e, buttonEvent) {
				fmt.Printf("New order received!")

				globalOrderAssignments := runHallRequestAlgorithm(e, buttonEvent)
				localAssignedHallOrdersCh <- globalOrderAssignments[e.GetID()]
				network.SendHallOrderRedistribution(globalOrderAssignments, e.GetID())
				
			} else {
				fmt.Println("Order already in queue, not sending to algorithm")
			}

		case heartBeat := <-updateWorldViewCh:

			e.UpdateWorldView(&heartBeat)
			fmt.Printf("Updated worldview with heartbeat from %d received to %d\n", heartBeat.GetID(), e.GetID())

			onUpdateWorldView(e)

			if !e.GetIsMaster() {
				fmt.Println("Switching to slave")
				break Loop
			}

		case peer := <-peerLostCh:
			fmt.Println("We lost peer ", peer)

			e.LooseConnectionToPeer(peer)

		}

		time.Sleep(10 * time.Millisecond)
	}

	//Maybe make onMasterSlaveChange-function
	e.SetIsMaster(false)
	e.UpdateMyBackup()
	go SlaveFsm(e, hallButtonCh, globalAssignedHallOrdersCh, localAssignedHallOrdersCh, updateWorldViewCh, peerLostCh)

}

// Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
// Idk om dette er en god løsning..
func SlaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, globalAssignedHallOrdersCh <-chan map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool,
	localAssignedHallOrdersCh chan<- [config.N_FLOORS][config.N_BUTTONS - 1]bool, updateWorldViewCh <-chan elevator.Backup, peerLostCh <-chan int) {

	fmt.Println("I am slave")
	fmt.Printf("Master is: %d \n", e.GetMasterID())

Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Give to masterHallOrderRequest
			network.SendHallOrder(buttonEvent, e.GetID(), e.GetMasterID())

		case heartBeat := <-updateWorldViewCh:

			e.UpdateWorldView(&heartBeat)
			fmt.Printf("Updated worldview with heartbeat from %d received to %d\n", heartBeat.GetID(), e.GetID())

			onUpdateWorldView(e)

			if e.GetIsMaster() {
				fmt.Println("Switching to master")
				break Loop
			}

		case peer := <-peerLostCh:
			fmt.Println("We lost peer ", peer)

			e.LooseConnectionToPeer(peer)

			e.TryUpdateIsMaster()
			if e.GetIsMaster() {
				fmt.Println("Switching to master")
				break Loop
			}
		case globalHallOrders := <-globalAssignedHallOrdersCh:

			localAssignedHallOrdersCh <- globalHallOrders[e.GetID()]

		}

		time.Sleep(10 * time.Millisecond)

	}

	e.SetIsMaster(true)
	e.UpdateMyBackup()
	go MasterFsm(e, hallButtonCh, globalAssignedHallOrdersCh, localAssignedHallOrdersCh, updateWorldViewCh, peerLostCh)

}

func onUpdateWorldView(e *elevator.Elevator) {

	e.TryUpdateIsMaster()

	//Also check motorstop
	//Also check other stuff

}
