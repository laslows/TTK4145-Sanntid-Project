package fsm

import (
	"Sanntid/src/events"
	"Sanntid/src/orders"
	"fmt"
)

func MasterSlaveFsm(hallButtonCh <-chan events.ButtonEvent, assignedOrderCh chan<- orders.Order) {
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			//Should decide here who takes the orker. For now it is just sent back to the fsm
			assignedOrderCh <- orders.New(buttonEvent.GetFloor(), (orders.OrderType)(buttonEvent.GetButton()))
			fmt.Printf("Should assign order: %v\n", buttonEvent)
		}
	}

}
