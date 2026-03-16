package orders

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"

	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

//TODO:fix name

type systemState struct {
	HallRequests [config.N_FLOORS][config.N_BUTTONS - 1]bool                 `json:"hallRequests"`
	States       map[string]elevatorState `json:"states"`
}

type elevatorState struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func createJSONDataForHallRequestAlgorithm(e *elevator.Elevator) string {
	states := make(map[string]elevatorState)

	/*hallRequests := make([][]bool, config.N_FLOORS)
	for i := range hallRequests {
		hallRequests[i] = make([]bool, config.N_BUTTONS-1)
	}

	globalHallRequests := e.GetGlobalRequests()
    for floor := 0; floor < config.N_FLOORS; floor++ {
        for btn := 0; btn < config.N_BUTTONS-1; btn++ {
            hallRequests[floor][btn] = globalHallRequests[floor][btn]
        }
    }*/

	hallRequests := e.GetGlobalRequests()

	worldView := e.GetWorldView()

	for _, backup := range worldView {

		if backup == nil {
			continue
		}

		if backup.GetBehaviour() != elevator.MotorStop && !backup.GetIsObstructed() && backup.GetConnectedToNetwork() {

			cabRequests := backup.GetCabRequests()

			states[strconv.Itoa(backup.GetID())] = elevatorState{
				Behaviour:   elevator.BehaviourToString(backup.GetBehaviour()),
				Floor:       backup.GetFloor(),
				Direction:   elevator.DirectionToString(backup.GetDirection()),
				CabRequests: make([]bool, config.N_FLOORS),
			}
			for i := 0; i < config.N_FLOORS; i++ {
                states[strconv.Itoa(backup.GetID())].CabRequests[i] = cabRequests[i]
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

func RunHallRequestAlgorithm(e *elevator.Elevator) map[int][config.N_FLOORS][config.N_BUTTONS - 1]bool {
	input := createJSONDataForHallRequestAlgorithm(e)
	cmd := exec.Command("./src/orders/hall_request_assigner/hall_request_assigner", "--input", input)
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

func CheckNewOrder(e *elevator.Elevator, hallOrder Order) bool {

	return !e.GetGlobalRequests()[hallOrder.GetFloor()][hallOrder.GetOrderType()]

}
