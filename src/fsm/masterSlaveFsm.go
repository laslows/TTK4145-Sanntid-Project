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

//TODO? get a snapshot from backup in here

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
			//elevatorStates := network.GetElevatorStates() // TODO: implement functinality in network

			// Ensure we include OUR own up-to-date state in the map
			worldViewMap := make(map[string]elevator.Elevator)
			for _, backup := range elev.GetWorldView() {
				if backup != nil {
					elev.UpdateMyBackup()
					//worldViewMap[backup.GetIP()] = *backup
				}
			}
			assignment := OptimalHallRequests(hallReqs, worldViewMap)
			id := elevator.GetIPandPortAsInt(elev.GetIP(), elev.GetPort())

			if id == elevator.GetIPandPortAsInt(elev.GetIP(), elev.GetPort()) {
				assignedOrderCh <- orders.New(f, btn)
			} else {
				//Give to slave
				network.SendHallOrder(orders.New(f, btn), elevator.GetIPandPortAsInt(elev.GetIP(), elev.GetPort()),
					id, network.HallOrderAssignment)
			}
			// Starting point: only send the orders assigned to THIS elevator into local FSM, TODO senere
			if myGrid, ok := assignment[elev.GetIP()]; ok {
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

	elev.SetIsMaster(false)
	elev.UpdateMyBackup()

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
	//fmt.Printf("Master is: %d \n", elev.GetMasterIP())

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
			network.SendHallOrder(orders.New(buttonEvent.GetFloor(), orders.OrderType(buttonEvent.GetButton())), elevator.GetIPandPortAsInt(elev.GetIP(), elev.GetPort()),
				elev.GetMasterID(), network.HallOrderRequest)

		}

		time.Sleep(10 * time.Millisecond)

	}

	elev.SetIsMaster(true)
	elev.UpdateMyBackup()

	go MasterFsm(elev, hallButtonCh, assignedOrderCh, changeMasterSlaveCh)
}
