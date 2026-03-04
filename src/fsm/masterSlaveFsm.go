package fsm

import (
	"Sanntid/src/elevator"
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

func MasterFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrderCh chan<- orders.Order,
	changeMasterSlaveCh <-chan bool) {
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

		case isMaster := <-changeMasterSlaveCh:
			if !isMaster {
				fmt.Println("Switching to slave")
				break Loop
			} else {
				fmt.Println("Should not be here, stay master")
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	//Maybe make onMasterSlaveChange-function
	e.SetIsMaster(false)
	e.UpdateMyBackup()
	go SlaveFsm(e, hallButtonCh, assignedOrderCh, changeMasterSlaveCh)

}

// Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
// Idk om dette er en god løsning..
func SlaveFsm(e *elevator.Elevator, hallButtonCh <-chan orders.Order, assignedOrderCh chan<- orders.Order, changeMasterSlaveCh <-chan bool) {

	fmt.Println("I am slave")
	fmt.Printf("Master is: %d \n", e.GetMasterID())

Loop:
	for {
		select {
		case isMaster := <-changeMasterSlaveCh:
			if isMaster {
				fmt.Println("Switching to master")
				break Loop
			} else {
				fmt.Println("Should not be here, stay slave")
			}
		case buttonEvent := <-hallButtonCh:
			//Give to master
			network.SendHallOrder(buttonEvent, e.GetID(), e.GetMasterID(), network.HallOrderRequest)

		}

		time.Sleep(10 * time.Millisecond)

	}

	e.SetIsMaster(true)
	e.UpdateMyBackup()
	go MasterFsm(e, hallButtonCh, assignedOrderCh, changeMasterSlaveCh)

}
