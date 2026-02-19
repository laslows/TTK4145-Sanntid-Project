package fsm

import (
	"Sanntid/src/events"
	"Sanntid/src/orders"
	"fmt"
)

func MasterFsm(hallButtonCh <-chan events.ButtonEvent, assignedOrderCh chan<- orders.Order, changeMasterSlaveCh <-chan bool) {
Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Should decide here who takes the orker. For now it is just sent back to the fsm
			assignedOrderCh <- orders.New(buttonEvent.GetFloor(), (orders.OrderType)(buttonEvent.GetButton()))
			fmt.Printf("Should assign order: %v\n", buttonEvent)
		case <-changeMasterSlaveCh:
			break Loop
		}
	}

	go SlaveFsm(/*Same inputs as master*/)

}

//Kanskje vi kan returne fra masterFsm om vi bli slave, og starte denne. Og så motsatt ??
//Idk om dette er en god løsning..
func SlaveFsm(/*Same inputs as master*/) {


}

