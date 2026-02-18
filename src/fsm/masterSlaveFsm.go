package fsm

import (
	"Sanntid/src/orders"
	"fmt"
)

func MasterSlaveFsm(newHallOrderCh <-chan orders.OrderType) {
	for {
		select {
		case order := <-newHallOrderCh:
			//Assign order
			fmt.Printf("Should assign order: %v\n", order)

		}
	}

}
