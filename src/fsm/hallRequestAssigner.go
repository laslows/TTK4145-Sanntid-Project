package fsm

import (
	"Sanntid/src/config"
	"Sanntid/src/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

//TODO:fix name

type SystemState struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       map[string]ElevatorState `json:"states"`
}

type ElevatorState struct {
	Behaviour   string `json:"behaviour"`
	Floor       int `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func createJSONDataForHallReqAlgorithm(e *elevator.Elevator) string {
	states := make(map[string]ElevatorState)

	hallRequests := make([][]bool, config.N_FLOORS)

	for i := range hallRequests {
		hallRequests[i] = make([]bool, config.N_BUTTONS-1)
	}

	worldView := e.GetWorldView()
	
	for _, backup := range worldView {
		fmt.Println(backup.GetConnectedToNetwork())
		fmt.Println(backup.GetBehaviour())
		if backup != nil {
			requests := backup.GetRequests()
			if backup.GetBehaviour() != elevator.MotorStop && backup.GetConnectedToNetwork() {
				fmt.Println(backup.GetID())
				states[strconv.Itoa(backup.GetID())] = ElevatorState{
					Behaviour:   elevator.BehaviourToString(backup.GetBehaviour()),
					Floor:       backup.GetFloor(),
					Direction:   elevator.DirectionToString(backup.GetDirection()),
					CabRequests: make([]bool, len(requests)),
				}
				for i, row := range requests {
					states[strconv.Itoa(backup.GetID())].CabRequests[i] = row[len(row)-1]
				}
			}
			for i := range hallRequests {
				for j := range hallRequests[i] {
					hallRequests[i][j] = hallRequests[i][j] || requests[i][j]
				}
			}
		}
	}

	println(len(states))
	system := SystemState{
		HallRequests: hallRequests,
		States:       states,
	}
	jsonData, err := json.Marshal(system)
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func runHallReqAlgorithm(e *elevator.Elevator) map[int][config.N_FLOORS][config.N_BUTTONS-1]bool {
	input := createJSONDataForHallReqAlgorithm(e)
	cmd := exec.Command("./src/fsm/hall_request_assigner/hall_request_assigner", "--input", input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Errorf("running hall request algorithm: %w; output: %s", err, string(out))
	}
	fmt.Print(string(out))
	var hallOrderAssignmentMap map[int][config.N_FLOORS][config.N_BUTTONS-1]bool

	err = json.Unmarshal(out, &hallOrderAssignmentMap)
	if err != nil {
		panic(err)
	}

	return hallOrderAssignmentMap
}
