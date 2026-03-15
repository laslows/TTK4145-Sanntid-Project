package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

//TODO:fix name

type systemState struct {
	HallRequests [][]bool                 `json:"hallRequests"`
	States       map[string]elevatorState `json:"states"`
}

type elevatorState struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func createJSONDataForHallRequestAlgorithm(e *elevator.Elevator, hallOrder *orders.Order) string {
	states := make(map[string]elevatorState)

	hallRequests := make([][]bool, config.N_FLOORS)

	for i := range hallRequests {
		hallRequests[i] = make([]bool, config.N_BUTTONS-1)
	}

	if hallOrder != nil {
		hallRequests[hallOrder.GetFloor()][hallOrder.GetOrderType()] = true
	}

	worldView := e.GetWorldView()

	for _, backup := range worldView {

		if backup == nil {
			continue
		}

		backupRequests := backup.GetRequests()
		if backup.GetBehaviour() != elevator.MotorStop && !backup.GetIsObstructed() && backup.GetConnectedToNetwork() {

			states[strconv.Itoa(backup.GetID())] = elevatorState{
				Behaviour:   elevator.BehaviourToString(backup.GetBehaviour()),
				Floor:       backup.GetFloor(),
				Direction:   elevator.DirectionToString(backup.GetDirection()),
				CabRequests: make([]bool, len(backupRequests)),
			}
			for i, row := range backupRequests {
				states[strconv.Itoa(backup.GetID())].CabRequests[i] = row[len(row)-1]
			}
		}

		for floor := range hallRequests {
			for button := range hallRequests[floor] {
				hallRequests[floor][button] = hallRequests[floor][button] || backupRequests[floor][button]
			}
		}
	}

	system := systemState{
		HallRequests: hallRequests,
		States:       states,
	}
	jsonData, err := json.Marshal(system)
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

func runHallRequestAlgorithm(e *elevator.Elevator, hallOrder *orders.Order) map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool {
	input := createJSONDataForHallRequestAlgorithm(e, hallOrder)
	cmd := exec.Command("./src/fsm/hall_request_assigner/hall_request_assigner", "--input", input)
	out, err := cmd.CombinedOutput()
	hallOrderAssignmentMap := make(map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool)

	if err != nil {
		fmt.Printf("running hall request algorithm failed: %v; output: %s\n", err, string(out))
		return hallOrderAssignmentMap
	}
	fmt.Print(string(out))

	err = json.Unmarshal(out, &hallOrderAssignmentMap)
	if err != nil {
		fmt.Printf("Json unmarshal failed: %v\n", err)
		return hallOrderAssignmentMap
	}

	for _, backup := range e.GetWorldView() {
		if backup != nil && (backup.GetIsObstructed() || backup.GetHasMotorstop()) {
			hallOrderAssignmentMap[backup.GetID()] = [config.N_FLOORS][config.N_BUTTONS - 1]bool{}
		}
	}

	return hallOrderAssignmentMap
}

func checkNewOrder(e *elevator.Elevator, hallOrder orders.Order) bool {
	//Check if order is already in queue, if not return true, else return false
	for _, backup := range e.GetWorldView() {
		if backup != nil {
			requests := backup.GetRequests()

			if requests[hallOrder.GetFloor()][hallOrder.GetOrderType()] {
				return false
			}
		}
	}
	return true

}
