package fsm

import (
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrderCh chan<- orders.Order,
	updateWorldViewCh <-chan elevator.Backup, peerLostCh <-chan int) {
Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Should decide here who takes the order. For now it is just sent back to the fsm
			chosenElevator := SuperOptimalOrderAssignmentAlgorithm()

			id := e.GetWorldView()[chosenElevator].GetID()

			if id == e.GetID() {
				assignedOrderCh <- buttonEvent
			} else {
				//Give to slave
				network.SendHallOrder(buttonEvent, e.GetID(),
					id, network.HallOrderAssignment)
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
	go SlaveFsm(e, hallButtonCh, assignedOrderCh, updateWorldViewCh, peerLostCh)

}

// Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
// Idk om dette er en god løsning..
func SlaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrderCh chan<- orders.Order, updateWorldViewCh <-chan elevator.Backup,
	peerLostCh <-chan int) {

	fmt.Println("I am slave")
	fmt.Printf("Master is: %d \n", e.GetMasterID())

Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Give to master
			network.SendHallOrder(buttonEvent, e.GetID(), e.GetMasterID(), network.HallOrderRequest)

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
		}

		time.Sleep(10 * time.Millisecond)

	}

	e.SetIsMaster(true)
	e.UpdateMyBackup()
	go MasterFsm(e, hallButtonCh, assignedOrderCh, updateWorldViewCh, peerLostCh)

}

func onUpdateWorldView(e *elevator.Elevator) {

	e.TryUpdateIsMaster()

	//Also check motorstop
	//Also check other stuff

}
