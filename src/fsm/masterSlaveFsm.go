package fsm

import (
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"Sanntid/src/config"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedHallOrdersCh chan<- map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool,
	updateWorldViewCh <-chan elevator.Backup, peerLostCh <-chan int) {
Loop:
	for {
		select {
		// case buttonEvent := <-hallButtonCh:
			
		// 	orderAssignments := runHallReqAlgorithm(e)
			
		// 	//Do stuff with own hall orders

			
		// 	network.SendHallOrderRedistribution(orderAssignments, e.GetID())

		case heartBeat := <-updateWorldViewCh:
			fmt.Println("master")
			fmt.Println(heartBeat)

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
	go SlaveFsm(e, hallButtonCh, assignedHallOrdersCh, updateWorldViewCh, peerLostCh)

}

// Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
// Idk om dette er en god løsning..
func SlaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedHallOrdersCh chan<- map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool, updateWorldViewCh <-chan elevator.Backup,
	peerLostCh <-chan int) {

	fmt.Println("I am slave")
	fmt.Printf("Master is: %d \n", e.GetMasterID())

Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Give to masterHallOrderRequest
			network.SendHallOrder(buttonEvent, e.GetID(), e.GetMasterID())

		case heartBeat := <-updateWorldViewCh:
			fmt.Println("slave")

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
		}

		time.Sleep(10 * time.Millisecond)

	}

	e.SetIsMaster(true)
	e.UpdateMyBackup()
	go MasterFsm(e, hallButtonCh, assignedHallOrdersCh, updateWorldViewCh, peerLostCh)

}

func onUpdateWorldView(e *elevator.Elevator) {

	e.TryUpdateIsMaster()

	//Also check motorstop
	//Also check other stuff

}
