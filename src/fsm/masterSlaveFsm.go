package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/events"
	"Sanntid/src/network"
	"Sanntid/src/orders"
	"fmt"
	"time"
)

func MasterFsm(
	elev *elevator.Elevator,
	hallButtonCh <-chan events.ButtonEvent,
	assignedOrderCh chan<- orders.Order,
	changeMasterSlaveCh <-chan bool,
) {
	fmt.Println("I am master")

	// Global hall request set (master-owned)
	hallReqs := make([][2]bool, config.N_FLOORS)

Loop:
	for {
		select {
		case buttonEvent := <-hallButtonCh:
			f := buttonEvent.GetFloor()
			btn := orders.OrderType(buttonEvent.GetButton())

			// Update hallReqs
			if f >= 0 && f < config.N_FLOORS {
				switch btn {
				case orders.HALL_UP:
					hallReqs[f][0] = true
				case orders.HALL_DOWN:
					hallReqs[f][1] = true
				default:
					// ignore (master fsm should only handle hall buttons)
				}
			}

			// ---- TO-DO: add function to run hall assigner ----
			elevatorStates := network.GetElevatorStates() // TODO: implement functinality in network

			// Ensure we include OUR own up-to-date state in the map
			myID := elev.GetPort() // starting point: use port as ID
			elevatorStates[myID] = *elev

			assignments := OptimalHallRequests(hallReqs, elevatorStates) //adds optimalhallrequetss

			// Starting point: only send the orders assigned to THIS elevator into local FSM, TODO senere
			if myGrid, ok := assignments[myID]; ok {
				for floor := 0; floor < len(myGrid); floor++ {
					if myGrid[floor][0] {
						assignedOrderCh <- orders.New(floor, orders.HALL_UP)
					}
					if myGrid[floor][1] {
						assignedOrderCh <- orders.New(floor, orders.HALL_DOWN)
					}
				}
			}

			// TODO (later):
			// - send assigned hall orders for other elevators over the network
			// - clear hallReqs when served
			// - avoid re-sending duplicates every time
			// -------------------------------

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

	go SlaveFsm(elev, hallButtonCh, assignedOrderCh, changeMasterSlaveCh)
}

// Slave forwards hall button events to master.
func SlaveFsm(
	elev *elevator.Elevator,
	hallButtonCh <-chan events.ButtonEvent,
	assignedOrderCh chan<- orders.Order,
	changeMasterSlaveCh <-chan bool,
) {
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
			// Forward to master
			network.SendHallOrderToMaster(
				orders.New(buttonEvent.GetFloor(), orders.OrderType(buttonEvent.GetButton())),
			)
		}

		time.Sleep(10 * time.Millisecond)
	}

	go MasterFsm(elev, hallButtonCh, assignedOrderCh, changeMasterSlaveCh)
}
