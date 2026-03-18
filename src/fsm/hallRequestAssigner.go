package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"Sanntid/src/orders"
	"encoding/json"
	"os/exec"
	"strconv"
)

//TODO:fix name

type systemState struct {
	M_hallRequests [config.N_FLOORS][config.N_BUTTONS - 1]bool `json:"hallRequests"`
	M_states       map[string]elevatorState                    `json:"states"`
}

type elevatorState struct {
	M_behaviour   string `json:"behaviour"`
	M_floor       int    `json:"floor"`
	M_direction   string `json:"direction"`
	M_cabRequests []bool `json:"cabRequests"`
}

func createJSONDataForHallRequestAlgorithm(e *elevator.Elevator, hallOrder *orders.Order) string {
	states := make(map[string]elevatorState)

	hallRequests := [config.N_FLOORS][config.N_BUTTONS - 1]bool{}

	if hallOrder != nil {
		hallRequests[hallOrder.GetFloor()][hallOrder.GetOrderType()] = true
	}

	worldView := e.GetWorldView()

	for _, backup := range worldView {

		if backup == nil {
			continue
		}

		backupRequests := backup.GetRequests()
		if backup.GetBehaviour() != elevator.MotorStop && !backup.GetIsObstructed() && backup.GetIsConnectedToNetwork() {

			states[strconv.Itoa(backup.GetID())] = elevatorState{
				M_behaviour:   elevator.BehaviourToString(backup.GetBehaviour()),
				M_floor:       backup.GetFloor(),
				M_direction:   elevator.DirectionToString(backup.GetDirection()),
				M_cabRequests: make([]bool, len(backupRequests)),
			}
			for i, row := range backupRequests {
				states[strconv.Itoa(backup.GetID())].M_cabRequests[i] = row[len(row)-1]
			}
		}

		for floor := range hallRequests {
			for button := range hallRequests[floor] {
				hallRequests[floor][button] = hallRequests[floor][button] || backupRequests[floor][button]
			}
		}
	}

	system := systemState{
		M_hallRequests: hallRequests,
		M_states:       states,
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
		return hallOrderAssignmentMap
	}

	err = json.Unmarshal(out, &hallOrderAssignmentMap)
	if err != nil {
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
