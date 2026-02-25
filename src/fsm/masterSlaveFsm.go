package fsm

import (
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

func MasterFsm(hallButtonCh <-chan orders.Order, assignedOrderCh chan<- orders.Order, changeMasterSlaveCh <-chan bool) {
Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Should decide here who takes the order. For now it is just sent back to the fsm
			assignedOrderCh <- buttonEvent
			fmt.Printf("Should assign order: %v\n", buttonEvent)

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

	go SlaveFsm(hallButtonCh, assignedOrderCh, changeMasterSlaveCh)

}

// Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
// Idk om dette er en god løsning..
func SlaveFsm(hallButtonCh <-chan orders.Order, assignedOrderCh chan<- orders.Order, changeMasterSlaveCh <-chan bool) {

	fmt.Println("I am slave")

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
			network.SendHallOrderToMaster(buttonEvent)

		}

		time.Sleep(10 * time.Millisecond)

	}

	go MasterFsm(hallButtonCh, assignedOrderCh, changeMasterSlaveCh)

}
